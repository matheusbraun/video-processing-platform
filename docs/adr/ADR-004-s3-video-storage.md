# ADR-004: AWS S3 para armazenamento de vídeos e frames

**Status:** Aceito
**Data:** 2026-02-01

---

## Contexto

O sistema precisa armazenar:
- **Vídeos originais**: arquivos de até 500 MB enviados pelos usuários
- **Frames extraídos**: potencialmente centenas de imagens JPEG por vídeo (1 frame/segundo)
- **Arquivos ZIP**: pacote compactado de todos os frames de um vídeo

Três alternativas foram consideradas:

1. **Disco local nos contêineres**: simples, sem custo de egress entre serviços. Problema: containers são efêmeros, escalar workers implica que cada instância teria apenas seu próprio disco, impossibilitando que um worker processe um vídeo enviado por outro nó.
2. **Volume NFS compartilhado**: resolve o problema de compartilhamento. Problema: o NFS se torna ponto único de falha, tem performance limitada para grandes arquivos e exige infraestrutura adicional para manter.
3. **Armazenamento de objetos (AWS S3)**: API unificada para upload/download de qualquer instância, durabilidade de 99,999999999% (11 noves), escala infinita, sem gerenciamento de capacidade.

---

## Decisão

Adotar **AWS S3** com **dois buckets separados** e acesso via URLs pré-assinadas para downloads:

**Estrutura de buckets:**
```
S3_UPLOADS_BUCKET (video-platform-uploads)
└── uploads/{video_id}/original.{ext}

S3_PROCESSED_BUCKET (video-platform-processed)
└── processed/{video_id}/
    ├── frame_0001.jpg
    ├── frame_0002.jpg
    └── {video_id}.zip
```

**Dois buckets** garantem que políticas de ciclo de vida (lifecycle) e permissões possam ser configuradas independentemente para cada tipo de dado.

**Upload de vídeos:** streaming multipart via AWS SDK v2 (`s3manager.Uploader`), sem carregar o arquivo inteiro em memória no API Gateway.

**Download de frames:** o Processing Worker baixa o vídeo original do bucket de uploads e envia os frames diretamente para o bucket de processados, sem passar pelo API Gateway.

**Downloads do usuário:** geração de **URLs pré-assinadas** com expiração de 15 minutos. O API Gateway solicita a URL ao S3 e retorna ao cliente — o cliente faz download direto do S3, sem tráfego passando pelos serviços da aplicação.

**Implementação:** `shared/pkg/storage/s3/` encapsula todas as operações S3 (upload, download, delete, presign) por trás de uma interface, permitindo que mocks sejam usados nos testes unitários.

---

## Consequências

**Positivas:**
- Todos os serviços (API Gateway, Processing Worker, Storage Service) acessam os mesmos arquivos independentemente de qual instância está executando — containers permanecem **stateless**.
- Escala automaticamente: sem preocupação com capacidade de disco ou IOPS.
- Buckets privados: nenhum arquivo é acessível publicamente; URLs pré-assinadas garantem acesso temporário e rastreável.
- Durabilidade e disponibilidade gerenciadas pela AWS.
- Downloads pesados (arquivos ZIP de centenas de MB) não sobrecarregam os serviços — vão direto do S3 para o cliente.

**Negativas:**
- Dependência de provedor de nuvem (vendor lock-in na AWS). Mitigável usando MinIO ou outro serviço S3-compatível em ambientes locais ou alternativos.
- Custo por GB armazenado e por requisição de PUT/GET. Mitigado pela política de retenção de 15 dias (Cleanup Cron).
- Latência de rede adicional comparada ao disco local — irrelevante para vídeos grandes onde o throughput é o gargalo, não a latência.

**Neutras:**
- A interface `S3Client` em `shared/pkg/storage/s3/` permite substituir a implementação AWS por MinIO para testes ou ambientes sem acesso à nuvem.
- A expiração das URLs pré-assinadas (15 min) pode ser ajustada via variável de ambiente se fluxos de download lentos exigirem mais tempo.
