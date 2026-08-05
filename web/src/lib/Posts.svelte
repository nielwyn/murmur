<script lang="ts">
    import { fade } from "svelte/transition";
    import { SvelteSet } from "svelte/reactivity";
    import { api, ApiError, type Post } from "./api";

    let { onGoToFeeds }: { onGoToFeeds: () => void } = $props();

    const loadedImages = new SvelteSet<string>();
    const failedImages = new SvelteSet<string>();

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

    let streamFoot: HTMLElement | undefined = $state();
    let posts: Post[] = $state([]);
    let filter: "all" | "unread" = $state("all");
    let hasFollows = $state(true);
    let loadingOlder = $state(false);
    let ended = $state(false);
    let loading = $state(true);
    let error = $state("");
    let requestId = 0;
    let showBackToTop = $state(false);

    // Server already filters by unread; this just hides a post instantly on mark-read.
    const visible = $derived(posts.filter((p) => filter === "all" || !p.read));
    // Fetched separately (not derived from `posts`) since `posts` only ever
    // holds the currently loaded page, not every unread post on the server.
    let unreadTotal = $state(0);
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

    async function setRead(post: Post, read: boolean) {
        const previous = post.read;
        if (previous === read) return;
        post.read = read;
        unreadTotal += read ? -1 : 1;
        try {
            if (read) {
                await api.markPostRead(post.id);
            } else {
                await api.markPostUnread(post.id);
            }
        } catch (e) {
            post.read = previous;
            unreadTotal += read ? 1 : -1;
            error =
                e instanceof ApiError
                    ? e.message
                    : "could not update read status";
        }
    }

    async function markAllRead() {
        const unread = posts.filter((p) => !p.read);
        unread.forEach((p) => (p.read = true));
        unreadTotal -= unread.length;
        try {
            await Promise.all(unread.map((p) => api.markPostRead(p.id)));
        } catch (e) {
            error =
                e instanceof ApiError ? e.message : "could not mark all read";
            await load(filter === "unread");
            await loadUnreadTotal();
        }
    }

    async function checkHasFollows() {
        try {
            hasFollows = (await api.listFollowing()).length > 0;
        } catch {}
    }

    async function loadUnreadTotal() {
        try {
            unreadTotal = await api.unreadCount();
        } catch {}
    }

    async function loadOlder() {
        if (loadingOlder || ended || posts.length === 0) return;
        const id = requestId;
        loadingOlder = true;
        error = "";
        try {
            const last = posts[posts.length - 1];
            const older = await api.listPosts({
                before: { id: last.id, published_at: last.published_at },
                unread: filter === "unread",
            });
            if (id !== requestId) return; // filter changed mid-flight
            if (older.length === 0) {
                ended = true;
            } else {
                posts.push(...older);
            }
        } catch (e) {
            if (id !== requestId) return;
            error =
                e instanceof ApiError
                    ? e.message
                    : "could not load older posts";
        } finally {
            if (id === requestId) loadingOlder = false;
        }
    }

    async function load(unreadOnly: boolean) {
        const id = ++requestId;
        loading = true;
        error = "";
        ended = false;
        try {
            const result = await api.listPosts({ unread: unreadOnly });
            if (id !== requestId) return; // a newer filter change superseded this
            posts = result;
        } catch (e) {
            if (id !== requestId) return;
            error = e instanceof ApiError ? e.message : "could not load posts";
        } finally {
            if (id === requestId) loading = false;
        }
    }

    checkHasFollows();
    loadUnreadTotal();
    $effect(() => {
        load(filter === "unread");
    });

    // Tracked as state (not calling loadOlder directly in the observer) so
    // the effect below keeps retrying if one page isn't enough to push the
    // marker back out of view.
    let isIntersecting = $state(false);

    $effect(() => {
        if (!streamFoot) return;
        const observer = new IntersectionObserver(
            (entries) => {
                isIntersecting = entries[0].isIntersecting;
            },
            { rootMargin: "600px 0px" },
        );
        observer.observe(streamFoot);
        return () => observer.disconnect();
    });

    // !error avoids retrying in a tight loop on sustained failure; the retry button below clears it.
    $effect(() => {
        if (
            isIntersecting &&
            !loadingOlder &&
            !ended &&
            !error &&
            posts.length > 0
        ) {
            loadOlder();
        }
    });

    function scrollToTop() {
        window.scrollTo({
            top: 0,
            behavior: matchMedia("(prefers-reduced-motion: reduce)").matches
                ? "auto"
                : "smooth",
        });
    }

    $effect(() => {
        const onScroll = () => {
            showBackToTop = window.scrollY > 600;
        };
        window.addEventListener("scroll", onScroll, { passive: true });
        return () => window.removeEventListener("scroll", onScroll);
    });
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
        {#if posts.length === 0 && !hasFollows}
            <span class="display">Nothing here yet.</span>
            <span class="section-label">
                follow a feed to start your stream
            </span>
            <button class="link-tab" onclick={onGoToFeeds}>
                → go to feeds
            </button>
        {:else if posts.length === 0}
            <span class="display">Nothing here yet.</span>
            <span class="section-label">
                no posts fetched for your feeds yet — check back soon
            </span>
        {:else}
            <span class="display">You're caught up.</span>
            <span class="section-label">
                nothing unread — the murmuration is quiet
            </span>
        {/if}
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
                                {post.feed_title} · {timeAgo(post.published_at)}
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
                        <div class="post-body">
                            {#if post.image_url && !failedImages.has(post.image_url)}
                                <div
                                    class="post-thumb-wrap"
                                    class:is-loaded={loadedImages.has(
                                        post.image_url,
                                    )}
                                >
                                    <img
                                        class="post-thumb"
                                        class:is-loaded={loadedImages.has(
                                            post.image_url,
                                        )}
                                        src={post.image_url}
                                        alt=""
                                        loading="lazy"
                                        decoding="async"
                                        onload={() =>
                                            loadedImages.add(post.image_url!)}
                                        onerror={() =>
                                            failedImages.add(post.image_url!)}
                                    />
                                </div>
                            {/if}
                            <div class="post-text">
                                <h3 class="post-title display">
                                    <a
                                        href={post.link}
                                        target="_blank"
                                        rel="noreferrer"
                                        onclick={() => setRead(post, true)}
                                        onauxclick={() => setRead(post, true)}
                                    >
                                        {post.title}
                                    </a>
                                </h3>
                                {#if post.description}
                                    <p class="post-desc">
                                        {@html post.description}
                                    </p>
                                {/if}
                            </div>
                        </div>
                    </li>
                {/each}
            </ol>
        </section>
    {/each}
{/if}

<div class="stream-foot" bind:this={streamFoot}>
    {#if ended}
        <span class="section-label end-mark">
            — beginning of your stream —
        </span>
    {:else if error}
        <button class="link-tab" onclick={loadOlder}>
            {error} — retry
        </button>
    {:else if loadingOlder}
        <span class="section-label loading-mark" aria-live="polite">
            loading older posts…
        </span>
    {/if}
</div>

{#if showBackToTop}
    <div class="back-to-top-layer">
        <div class="back-to-top-column">
            <button
                class="back-to-top"
                onclick={scrollToTop}
                aria-label="back to top"
                transition:fade={{ duration: fadeMs }}
            >
                <span class="back-to-top-arrow" aria-hidden="true">↑</span>
                top
            </button>
        </div>
    </div>
{/if}

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

    .post-body {
        display: flex;
        gap: 1rem;
        align-items: flex-start;
    }

    .post-thumb-wrap {
        position: relative;
        height: 5.5rem;
        width: 9.75rem;
        flex-shrink: 0;
        border-radius: 4px;
        overflow: hidden;
        background: var(--paper-raised);
    }

    .post-thumb-wrap::before {
        content: "";
        position: absolute;
        inset: 0;
        background: linear-gradient(
            90deg,
            var(--paper-raised) 25%,
            var(--rule) 50%,
            var(--paper-raised) 75%
        );
        background-size: 200% 100%;
        animation: thumb-shimmer 1.4s ease-in-out infinite;
        opacity: 1;
        transition: opacity 0.2s ease;
    }

    .post-thumb-wrap.is-loaded::before {
        opacity: 0;
    }

    .post-thumb {
        position: absolute;
        inset: 0;
        width: 100%;
        height: 100%;
        object-fit: contain;
        opacity: 0;
        transition: opacity 0.2s ease;
    }

    .post-thumb.is-loaded {
        opacity: 1;
    }

    @keyframes thumb-shimmer {
        0% {
            background-position: 200% 0;
        }
        100% {
            background-position: -200% 0;
        }
    }

    .post-text {
        min-width: 0;
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

    .back-to-top-layer {
        position: fixed;
        inset: 0;
        pointer-events: none;
        z-index: 10;
    }

    /* Keep in sync with main's max-width in App.svelte. */
    .back-to-top-column {
        position: relative;
        height: 100%;
        max-width: 42rem;
        margin-inline: auto;
    }

    .back-to-top {
        all: unset;
        position: absolute;
        right: 1.5rem;
        bottom: 1.5rem;
        display: flex;
        align-items: center;
        gap: 0.4rem;
        pointer-events: auto;
        background: var(--accent);
        border: 1px solid var(--accent);
        border-radius: 999px;
        padding: 0.5rem 1.1rem 0.5rem 0.9rem;
        font-family: var(--font-mono);
        font-size: 0.7rem;
        font-weight: 500;
        text-transform: uppercase;
        letter-spacing: 0.14em;
        color: var(--accent-ink);
        cursor: pointer;
        box-shadow: 0 3px 14px rgba(0, 0, 0, 0.25);
        transition:
            background-color 0.15s ease,
            box-shadow 0.15s ease;
    }

    .back-to-top-arrow {
        font-size: 1.25rem;
        line-height: 1;
    }

    .back-to-top:hover {
        background: var(--pico-primary-hover);
        border-color: var(--pico-primary-hover);
        box-shadow: 0 4px 18px rgba(0, 0, 0, 0.3);
    }

    .back-to-top:focus-visible {
        outline: 2px solid var(--accent);
        outline-offset: 2px;
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
        .read-toggle,
        .back-to-top {
            transition: none;
        }

        .post-thumb-wrap::before {
            animation: none;
        }
    }
</style>
