// Typed client for the murmur API. Auth is an httpOnly cookie, so there is
// no token handling here — the browser sends it automatically.

export interface User {
  id: string;
  username: string;
  email: string;
}

export interface Feed {
  id: string;
  name: string;
  url: string;
  creator_name?: string;
}

export interface Follow {
  feed_id: string;
  feed_name?: string;
  feed_url?: string;
}

export interface Post {
  id: string;
  feed_name: string;
  published_at?: string;
  title: string;
  url: string;
  description?: string;
  read?: boolean;
}

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    headers: init?.body ? { "Content-Type": "application/json" } : undefined,
    ...init,
  });

  if (!res.ok) {
    let message = res.statusText;
    try {
      message = ((await res.json()) as { error: string }).error;
    } catch {
      // Not JSON; keep the status text.
    }
    throw new ApiError(res.status, message);
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export const api = {
  register: (username: string, email: string, password: string) =>
    request<User>("/api/register", {
      method: "POST",
      body: JSON.stringify({ username, email, password }),
    }),

  login: (username: string, password: string) =>
    request<User>("/api/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    }),

  logout: () => request<void>("/api/logout", { method: "POST" }),

  me: () => request<User>("/api/me"),

  listFeeds: () => request<Feed[]>("/api/feeds"),

  createFeed: (name: string, url: string) =>
    request<Feed>("/api/feeds", {
      method: "POST",
      body: JSON.stringify({ name, url }),
    }),

  listFollowing: () => request<Follow[]>("/api/feeds/following"),

  followFeed: (feedId: string) =>
    request<Follow>(`/api/feeds/${feedId}/follow`, { method: "POST" }),

  unfollowFeed: (feedId: string) =>
    request<void>(`/api/feeds/${feedId}/follow`, { method: "DELETE" }),

  listPosts: () => request<Post[]>("/api/posts/"),
};
