import ky from 'ky';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export const api = ky.create({
  prefixUrl: API_BASE_URL,
  timeout: 30000,
  hooks: {
    beforeRequest: [
      (request) => {
        const token = localStorage.getItem('access_token');
        if (token) {
          request.headers.set('Authorization', `Bearer ${token}`);
        }
      },
    ],
    afterResponse: [
      async (_request, _options, response) => {
        if (response.status === 401) {
          const refreshToken = localStorage.getItem('refresh_token');
          if (refreshToken) {
            try {
              const result = await ky
                .post(`${API_BASE_URL}/api/v1/auth/refresh`, {
                  json: { refresh_token: refreshToken },
                })
                .json<{
                  data: { access_token: string; refresh_token: string };
                }>();

              localStorage.setItem('access_token', result.data.access_token);
              localStorage.setItem('refresh_token', result.data.refresh_token);

              return ky(_request.url, {
                ..._options,
                headers: {
                  ..._options.headers,
                  Authorization: `Bearer ${result.data.access_token}`,
                },
              });
            } catch {
              localStorage.removeItem('access_token');
              localStorage.removeItem('refresh_token');
              window.location.href = '/login';
            }
          } else {
            window.location.href = '/login';
          }
        }
        return response;
      },
    ],
  },
});
