export const api = {
  getAuthToken() {
    if (typeof window !== "undefined") {
      return localStorage.getItem("auth_token") || "";
    }
    return "";
  },

  async request(endpoint: string, options: RequestInit = {}) {
    const token = this.getAuthToken();
    const headers: HeadersInit = {
      "Content-Type": "application/json",
      ...options.headers,
    };

    if (token) {
      headers["Authorization"] = `Bearer ${token}`; // Seamless backend authorization parsing
    }

    // Tunnel traffic using native proxy rewrites
    const res = await fetch(`/api${endpoint}`, {
      ...options,
      headers,
    });

    if (!res.ok) {
      const errText = await res.text();
      throw new Error(errText || "Unexpected Generic API Error");
    }

    if (res.status === 204) return null;
    const text = await res.text();
    return text ? JSON.parse(text) : null;
  }
}
