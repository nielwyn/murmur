<script lang="ts">
    import { fade } from "svelte/transition";
    import { api, ApiError, type Post } from "./api";

    const rtf = new Intl.RelativeTimeFormat("en", { numeric: "auto" });

    function dayLabel(iso?: string): string {
        if (!iso) return "undated";
        const date = new Date(iso);
        const startOfDay = (d: Date) =>
            new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime();
        const diffDays = Math.round(
            (startOfDay(new Date()) - startOfDay(date)) / 86400000,
        );
        if (diffDays === 0) return "today";
        if (diffDays === 1) return "yesterday";
        return date
            .toLocaleDateString("en-US", {
                weekday: "long",
                month: "long",
                day: "numeric",
            })
            .toLowerCase();
    }

    function timeAgo(iso?: string): string {
        if (!iso) return "";
        const diffSec = (new Date(iso).getTime() - Date.now()) / 1000;
        const units: [Intl.RelativeTimeFormatUnit, number][] = [
            ["year", 31536000],
            ["month", 2592000],
            ["day", 86400],
            ["hour", 3600],
            ["minute", 60],
            ["second", 1],
        ];
        for (const [unit, secs] of units) {
            if (Math.abs(diffSec) >= secs || unit === "second") {
                return rtf.format(Math.round(diffSec / secs), unit);
            }
        }
        return "";
    }

    // One extra "page" so the pagination button has something to load.
    const OLDER_PAGE: Post[] = [
        {
            id: "13",
            feed_name: "the go blog",
            published_at: "2026-07-13T17:25:00Z",
            title: "Profile-guided optimization, two years in",
            url: "https://go.dev/blog/",
            description:
                "What PGO has delivered since 1.21, real-world numbers from large deployments, and how to start collecting profiles today.",
            read: true,
        },
        {
            id: "14",
            feed_name: "hacker news",
            published_at: "2026-07-13T10:58:00Z",
            title: "Ask HN: What's your self-hosting stack in 2026?",
            url: "https://news.ycombinator.com/",
            description:
                "Raspberry Pis, mini PCs, and old laptops. Docker Compose still rules, and everyone has an RSS reader in the list somewhere.",
            read: true,
        },
    ];

    let posts: Post[] = $state([]);
    let filter: "all" | "unread" = $state("all");
    let olderLoaded = $state(false);
    let ended = $state(false);
    let loading = $state(true);
    let error = $state("");

    const visible = $derived(posts.filter((p) => filter === "all" || !p.read));
    const unreadTotal = $derived(posts.filter((p) => !p.read).length);
    const groups = $derived.by(() => {
        const days: { day: string; items: Post[] }[] = [];
        for (const p of visible) {
            const label = dayLabel(p.published_at);
            let g = days.find((d) => d.day === label);
            if (!g) {
                g = { day: label, items: [] };
                days.push(g);
            }
            g.items.push(p);
        }
        return days;
    });

    const fadeMs = matchMedia("(prefers-reduced-motion: reduce)").matches
        ? 0
        : 200;

    function setRead(post: Post, read: boolean) {
        post.read = read;
    }

    function markAllRead() {
        posts.forEach((p) => (p.read = true));
    }

    function loadOlder() {
        if (olderLoaded) {
            ended = true;
        } else {
            posts.push(...OLDER_PAGE);
            olderLoaded = true;
        }
    }

    async function load() {
        loading = true;
        error = "";
        try {
            posts = await api.listPosts();
        } catch (e) {
            error = e instanceof ApiError ? e.message : "could not load posts";
        } finally {
            loading = false;
        }
    }

    load();
</script>

<div class="toolbar">
    <div class="filter" role="tablist" aria-label="filter posts">
        <button
            class="link-tab"
            role="tab"
            aria-selected={filter === "all"}
            onclick={() => (filter = "all")}
        >
            all
        </button>
        <button
            class="link-tab"
            role="tab"
            aria-selected={filter === "unread"}
            onclick={() => (filter = "unread")}
        >
            unread{unreadTotal > 0 ? ` (${unreadTotal})` : ""}
        </button>
    </div>
    <button class="link-tab" onclick={markAllRead}>mark all read</button>
</div>
<hr class="rule-heavy" />

{#if visible.length === 0}
    <div class="empty">
        <span class="display">You're caught up.</span>
        <span class="section-label">
            nothing unread — the murmuration is quiet
        </span>
    </div>
{:else}
    {#each groups as group (group.day)}
        {@const groupUnread = group.items.filter((p) => !p.read).length}
        <section class="day">
            <div class="day-head">
                <h2 class="section-label">{group.day}</h2>
                <span
                    class="section-label count"
                    class:has-unread={groupUnread > 0}
                >
                    {groupUnread > 0 ? `${groupUnread} unread` : "all read"}
                </span>
            </div>
            <ol class="posts">
                {#each group.items as post (post.id)}
                    <li
                        class="post"
                        class:is-read={post.read}
                        out:fade={{ duration: fadeMs }}
                    >
                        <div class="post-meta">
                            <span class="post-source section-label">
                                {post.feed_name} · {timeAgo(post.published_at)}
                            </span>
                            <button
                                class="read-toggle"
                                aria-label={post.read
                                    ? "mark as unread"
                                    : "mark as read"}
                                onclick={() => setRead(post, !post.read)}
                            >
                                {post.read ? "unread?" : "mark read"}
                            </button>
                        </div>
                        <h3 class="post-title display">
                            <a
                                href={post.url}
                                target="_blank"
                                rel="noreferrer"
                                onclick={() => setRead(post, true)}
                            >
                                {post.title}
                            </a>
                        </h3>
                        {#if post.description}
                            <p class="post-desc">{@html post.description}</p>
                        {/if}
                    </li>
                {/each}
            </ol>
        </section>
    {/each}
{/if}

<div class="stream-foot">
    {#if ended}
        <span class="section-label end-mark">
            — beginning of your stream ∎ —
        </span>
    {:else}
        <button class="link-tab" onclick={loadOlder}>↓ older posts</button>
    {/if}
</div>

<style>
    .toolbar {
        display: flex;
        justify-content: space-between;
        align-items: baseline;
        gap: 1rem;
        margin-bottom: 0.5rem;
    }

    .filter {
        display: flex;
        gap: 0.75rem;
        align-items: baseline;
    }

    .link-tab {
        all: unset;
        cursor: pointer;
        font-family: var(--font-mono);
        font-size: 0.7rem;
        font-weight: 500;
        text-transform: uppercase;
        letter-spacing: 0.18em;
        color: var(--ink-faint);
        padding-bottom: 2px;
        border-bottom: 1px solid transparent;
        transition:
            color 0.15s ease,
            border-color 0.15s ease;
    }

    .link-tab:hover {
        color: var(--accent);
    }

    .link-tab[aria-selected="true"] {
        color: var(--ink);
        border-bottom-color: var(--accent);
    }

    .link-tab:focus-visible {
        outline: 2px solid var(--accent);
        outline-offset: 3px;
    }

    .day {
        margin-top: 2rem;
    }

    .day-head {
        display: flex;
        justify-content: space-between;
        align-items: baseline;
        gap: 1rem;
        margin-bottom: 0.4rem;
    }

    .day-head h2 {
        margin: 0;
    }

    .count.has-unread {
        color: var(--accent);
    }

    ol.posts {
        list-style: none;
        margin: 0;
        padding: 0;
    }

    .post {
        padding: 1.05rem 0;
        border-bottom: 1px solid var(--rule);
    }

    .post-meta {
        display: flex;
        justify-content: space-between;
        align-items: baseline;
        gap: 1rem;
        margin-bottom: 0.35rem;
    }

    .post-source {
        display: inline-flex;
        align-items: baseline;
        gap: 0.5rem;
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    /* unread marker: one quiet oxblood dot, nothing louder */
    .post-source::before {
        content: "";
        width: 6px;
        height: 6px;
        border-radius: 50%;
        background: var(--accent);
        flex-shrink: 0;
        transform: translateY(-1px);
    }

    .post.is-read .post-source::before {
        background: transparent;
        border: 1px solid var(--rule);
    }

    .post-title {
        margin: 0;
        font-size: 1.3rem;
        line-height: 1.25;
        font-weight: 600;
        text-wrap: balance;
        font-variation-settings:
            "opsz" 40,
            "SOFT" 40,
            "WONK" 0;
    }

    .post-title a {
        color: var(--ink);
    }

    .post-title a:hover {
        color: var(--accent);
    }

    .post.is-read .post-title {
        font-weight: 450;
    }

    .post.is-read .post-title a {
        color: var(--ink-faint);
    }

    .post-desc {
        margin: 0.35rem 0 0;
        font-size: 0.9rem;
        max-width: 60ch;
        display: -webkit-box;
        -webkit-line-clamp: 2;
        line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }

    .post.is-read .post-desc {
        color: var(--ink-faint);
    }

    /* read toggle: a checkbox in newspaper clothes */
    .read-toggle {
        all: unset;
        cursor: pointer;
        font-family: var(--font-mono);
        font-size: 0.7rem;
        letter-spacing: 0.14em;
        text-transform: uppercase;
        color: var(--ink-faint);
        flex-shrink: 0;
        padding-bottom: 1px;
        border-bottom: 1px solid transparent;
        transition:
            color 0.15s ease,
            border-color 0.15s ease;
    }

    .read-toggle:hover {
        color: var(--accent);
        border-bottom-color: currentColor;
    }

    .read-toggle:focus-visible {
        outline: 2px solid var(--accent);
        outline-offset: 3px;
    }

    .stream-foot {
        margin-top: 2.5rem;
        padding-bottom: 2rem;
        text-align: center;
    }

    .end-mark {
        font-family: var(--font-display);
        letter-spacing: 0.3em;
    }

    .empty {
        display: flex;
        flex-direction: column;
        gap: 0.4rem;
        padding: 3rem 0 1rem;
        text-align: center;
    }

    .empty .display {
        font-size: 1.4rem;
    }

    @media (prefers-reduced-motion: reduce) {
        .link-tab,
        .read-toggle {
            transition: none;
        }
    }
</style>
