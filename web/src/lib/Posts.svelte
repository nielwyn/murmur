<script lang="ts">
    import { fade } from "svelte/transition";

    // Replace with GET /api/posts (+ POST/DELETE
    // /api/posts/{id}/read) and delete these constants.
    interface Post {
        id: number;
        feed: string;
        day: string;
        time: string;
        title: string;
        url: string;
        desc: string;
        read: boolean;
    }

    const DUMMY_POSTS: Post[] = [
        {
            id: 1,
            feed: "the go blog",
            day: "today",
            time: "2h ago",
            title: "Go 1.27 is released",
            url: "https://go.dev/blog/",
            desc: "Go 1.27 brings faster build times, new iterator helpers in the slices package, and improvements to the garbage collector on arm64.",
            read: false,
        },
        {
            id: 2,
            feed: "hacker news",
            day: "today",
            time: "3h ago",
            title: "SQLite as an application file format (2014)",
            url: "https://news.ycombinator.com/",
            desc: "Instead of inventing a custom binary format, treat a SQLite database file as your document format — transactional, queryable, and portable.",
            read: false,
        },
        {
            id: 3,
            feed: "julia evans",
            day: "today",
            time: "5h ago",
            title: "Notes on debugging DNS",
            url: "https://jvns.ca/",
            desc: "Some ways DNS can fail quietly, why the resolver cache is usually the culprit, and the three dig commands that answer 90% of questions.",
            read: false,
        },
        {
            id: 4,
            feed: "boot.dev blog",
            day: "today",
            time: "7h ago",
            title: "The bcrypt cost factor you actually want in 2026",
            url: "https://blog.boot.dev/",
            desc: "Password hashing needs to be slow on purpose. Here's how to pick a cost factor that survives modern GPUs without making login feel broken.",
            read: true,
        },
        {
            id: 5,
            feed: "hacker news",
            day: "today",
            time: "9h ago",
            title: "Show HN: I built an RSS reader that runs on a Raspberry Pi",
            url: "https://news.ycombinator.com/",
            desc: "One Go binary, embedded frontend, Postgres in a container. Fetches feeds with a bounded worker pool so the Pi never breaks a sweat.",
            read: false,
        },
        {
            id: 6,
            feed: "arch linux news",
            day: "yesterday",
            time: "21:40",
            title: "linux-firmware ≥ 20260701 requires manual intervention",
            url: "https://archlinux.org/news/",
            desc: "The package split changed file ownership. Run the documented pacman command before your next system upgrade to avoid a conflict.",
            read: true,
        },
        {
            id: 7,
            feed: "the go blog",
            day: "yesterday",
            time: "16:05",
            title: "Robust generic collections with iter.Seq",
            url: "https://go.dev/blog/",
            desc: "Patterns for writing collections that expose iterators instead of slices — and why returning iter.Seq keeps your API flexible.",
            read: false,
        },
        {
            id: 8,
            feed: "hacker news",
            day: "yesterday",
            time: "11:12",
            title: "The fan-out/fan-in pattern, benchmarked",
            url: "https://news.ycombinator.com/",
            desc: "How many workers is too many? Measuring throughput of a bounded worker pool against unbounded goroutines across I/O-heavy workloads.",
            read: true,
        },
        {
            id: 9,
            feed: "julia evans",
            day: "yesterday",
            time: "09:14",
            title: "New zine: How Containers Work!",
            url: "https://jvns.ca/",
            desc: "Namespaces, cgroups, and overlay filesystems explained with drawings — everything a container is, in 28 pages.",
            read: true,
        },
        {
            id: 10,
            feed: "boot.dev blog",
            day: "monday, july 14",
            time: "19:30",
            title: "Goroutines vs. threads: what actually happens in the scheduler",
            url: "https://blog.boot.dev/",
            desc: "M:N scheduling in plain words: why goroutines are cheap, what parks them, and what GOMAXPROCS really controls.",
            read: false,
        },
        {
            id: 11,
            feed: "hacker news",
            day: "monday, july 14",
            time: "14:02",
            title: "Miniflux 3.0 released",
            url: "https://news.ycombinator.com/",
            desc: "The minimalist feed reader gets keyboard-first navigation, a smaller Docker image, and OPML import improvements.",
            read: true,
        },
        {
            id: 12,
            feed: "arch linux news",
            day: "monday, july 14",
            time: "08:47",
            title: "PostgreSQL 18 moves to extra",
            url: "https://archlinux.org/news/",
            desc: "The new major version is now in the main repos. Remember: pg_upgrade before the old binaries are gone.",
            read: true,
        },
    ];

    // One extra "page" so the pagination button has something to load.
    const OLDER_PAGE: Post[] = [
        {
            id: 13,
            feed: "the go blog",
            day: "sunday, july 13",
            time: "17:25",
            title: "Profile-guided optimization, two years in",
            url: "https://go.dev/blog/",
            desc: "What PGO has delivered since 1.21, real-world numbers from large deployments, and how to start collecting profiles today.",
            read: true,
        },
        {
            id: 14,
            feed: "hacker news",
            day: "sunday, july 13",
            time: "10:58",
            title: "Ask HN: What's your self-hosting stack in 2026?",
            url: "https://news.ycombinator.com/",
            desc: "Raspberry Pis, mini PCs, and old laptops. Docker Compose still rules, and everyone has an RSS reader in the list somewhere.",
            read: true,
        },
    ];

    let posts: Post[] = $state([...DUMMY_POSTS]);
    let filter: "all" | "unread" = $state("all");
    let olderLoaded = $state(false);
    let ended = $state(false);

    const visible = $derived(posts.filter((p) => filter === "all" || !p.read));
    const unreadTotal = $derived(posts.filter((p) => !p.read).length);
    const groups = $derived.by(() => {
        const days: { day: string; items: Post[] }[] = [];
        for (const p of visible) {
            let g = days.find((d) => d.day === p.day);
            if (!g) {
                g = { day: p.day, items: [] };
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
                                {post.feed} · {post.time}
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
                            <!-- Opening an article marks it read. -->
                            <a
                                href={post.url}
                                target="_blank"
                                rel="noreferrer"
                                onclick={() => setRead(post, true)}
                            >
                                {post.title}
                            </a>
                        </h3>
                        {#if post.desc}
                            <p class="post-desc">{post.desc}</p>
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
