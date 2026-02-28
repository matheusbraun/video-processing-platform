# ADR-003: JWT com blacklist via Redis para gerenciamento de sessão

**Status:** Aceito
**Data:** 2026-02-01

---

## Contexto

A plataforma precisa autenticar usuários em múltiplos serviços (API Gateway, futuras expansões) de forma eficiente. Três abordagens foram consideradas:

1. **Sessões com estado no servidor (session cookies)**: o servidor mantém sessões em banco ou cache; escalabilidade requer sticky sessions ou session store compartilhada.
2. **JWT puro (stateless)**: tokens autocontidos validados localmente; sem necessidade de consulta ao banco em cada requisição. Problema: não há como revogar um token antes da expiração.
3. **JWT + refresh tokens + blacklist Redis**: tokens de acesso de curta duração com renovação via refresh token; logout efetivo via blacklist em cache.

A opção 1 exige que todos os serviços compartilhem a mesma session store, introduzindo acoplamento e ponto único de falha. A opção 2 não suporta logout real — um token roubado continuaria válido até expirar.

---

## Decisão

Adotar **JWT com dois tipos de token** e **blacklist de revogação via Redis**:

**Access token:**
- Algoritmo: HS256 com segredo configurável via `JWT_SECRET`
- Expiração: **15 minutos** (curta o suficiente para limitar o dano em caso de vazamento)
- Contém: `user_id`, `username`, `exp`, `jti` (JWT ID único)
- Validado **localmente** pelo middleware JWT do API Gateway — sem chamada de rede ao Auth Service

**Refresh token:**
- String opaca (UUID v4), armazenada em `auth.refresh_tokens` no PostgreSQL
- Expiração: **7 dias**
- Usado exclusivamente no endpoint `POST /refresh` do Auth Service para emitir novos access tokens

**Logout e revogação:**
- No logout, o refresh token é excluído do PostgreSQL E o `jti` do access token é inserido na blacklist Redis com TTL de 15 minutos (mesmo tempo de expiração do access token)
- O middleware JWT verifica a blacklist Redis antes de aceitar um access token
- Após os 15 minutos, o token expira naturalmente e a entrada Redis é removida automaticamente pelo TTL

**Validação JWT no middleware (shared/pkg/auth/jwt/):**
```
1. Extrai Bearer token do header Authorization
2. Valida assinatura com JWT_SECRET
3. Verifica campo exp (expiração)
4. Consulta Redis: EXISTS blacklist:{jti}
5. Extrai claims (user_id, username) e injeta no contexto da requisição
```

---

## Consequências

**Positivas:**
- Logout funciona de verdade: tokens revogados são rejeitados imediatamente (até o TTL expirar).
- Serviços validam tokens localmente (sem chamada de rede ao Auth Service por requisição), mantendo latência baixa.
- O sistema escala horizontalmente: qualquer instância do API Gateway acessa a mesma blacklist Redis.
- Refresh tokens de longa duração melhoram a UX (usuário não precisa fazer login frequentemente).

**Negativas:**
- Redis se torna dependência crítica para autenticação. Se Redis ficar indisponível, o logout não funciona e tokens potencialmente revogados são aceitos.
- Janela de revogação de até 15 minutos para access tokens caso o Redis não registre a entrada na blacklist (falha parcial).
- Adiciona uma consulta Redis por requisição autenticada.

**Neutras:**
- O padrão de tokens de curta duração é uma prática amplamente reconhecida para APIs REST.
- A configuração de expiração é externalizada via variáveis de ambiente (`JWT_ACCESS_EXPIRY`, `JWT_REFRESH_EXPIRY`), permitindo ajuste sem recompilação.
