// Typed client for the murmur API. Auth is an httpOnly cookie, so there is
// no token handling here — the browser sends it automatically.

export interface User {
  id: string;
  username: string;
  email: string;
}

export interface Feed {
  id: string;
  title: string;
  link: string;
  creator_name?: string;
}

export interface Follow {
  feed_id: string;
  feed_title?: string;
  feed_link?: string;
}

export interface Post {
  id: string;
  feed_title: string;
  published_at?: string;
  title: string;
  link: string;
  description?: string;
  read: boolean;
  image_url?: string;
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

  createFeed: (link: string) =>
    request<Feed>("/api/feeds", {
      method: "POST",
      body: JSON.stringify({ link }),
    }),

  listFollowing: () => request<Follow[]>("/api/feeds/following"),

  followFeed: (feedId: string) =>
    request<Follow>(`/api/feeds/${feedId}/follow`, { method: "POST" }),

  unfollowFeed: (feedId: string) =>
    request<void>(`/api/feeds/${feedId}/follow`, { method: "DELETE" }),

  listPosts: (opts?: {
    before?: { id: string; published_at?: string };
    unread?: boolean;
    feedTitle?: string;
  }) => {
    const params = new URLSearchParams();
    if (opts?.before) {
      params.set("before_id", opts.before.id);
      if (opts.before.published_at) {
        params.set("before_published_at", opts.before.published_at);
      }
    }
    if (opts?.unread) {
      params.set("unread", "true");
    }
    if (opts?.feedTitle) {
      params.set("feed", opts.feedTitle);
    }
    const qs = params.toString();
    return request<Post[]>(`/api/posts/${qs ? `?${qs}` : ""}`);
  },

  unreadCount: () =>
    request<{ count: number }>("/api/posts/unread-count").then(
      (r) => r.count,
    ),

  markPostRead: (postId: string) =>
    request<void>(`/api/posts/${postId}/read`, { method: "POST" }),

  markPostUnread: (postId: string) =>
    request<void>(`/api/posts/${postId}/read`, { method: "DELETE" }),
};
