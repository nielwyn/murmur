<script lang="ts">
    // DUMMY DATA — this page is design-first while GET /api/feeds/status
    // doesn't exist yet (week 6: RWMutex status map + rate limiter + this
    // endpoint). Replace with a real fetch and delete these constants.
    type FeedStatus = "ok" | "error" | "pending";
    // transient: probably temporary, will likely resolve on its own —
    // notFound: the feed itself is gone, retrying won't help.
    type ErrorKind = "transient" | "notFound";

    interface Feed {
        id: number;
        name: string;
        status: FeedStatus;
        checkedAgo?: string;
        newPosts?: number;
        errorKind?: ErrorKind;
        // Raw detail for anyone who wants it — shown as a hover tooltip,
        // never as the primary line (see the "why" thread on this page).
        errorDetail?: string;
        nextIn?: string;
        addedAgo?: string;
    }

    const DUMMY_FEEDS: Feed[] = [
        {
            id: 1,
            name: "a broken blog",
            status: "error",
            checkedAgo: "6m ago",
            errorKind: "transient",
            errorDetail: "connection timed out after 10s",
            nextIn: "54m",
        },
        {
            id: 2,
            name: "old tech digest",
            status: "error",
            checkedAgo: "1d ago",
            errorKind: "notFound",
            errorDetail: "404 not found",
        },
        {
            id: 3,
            name: "weekly rust news",
            status: "pending",
            addedAgo: "2m ago",
        },
        {
            id: 4,
            name: "hacker news",
            status: "ok",
            checkedAgo: "2m ago",
            newPosts: 5,
            nextIn: "13m",
        },
        {
            id: 5,
            name: "julia evans",
            status: "ok",
            checkedAgo: "34m ago",
            newPosts: 1,
            nextIn: "26m",
        },
        {
            id: 6,
            name: "the go blog",
            status: "ok",
            checkedAgo: "18m ago",
            newPosts: 0,
            nextIn: "42m",
        },
        {
            id: 7,
            name: "boot.dev blog",
            status: "ok",
            checkedAgo: "3h ago",
            newPosts: 2,
            nextIn: "3h",
        },
        {
            id: 8,
            name: "arch linux news",
            status: "ok",
            checkedAgo: "9h ago",
            newPosts: 0,
            nextIn: "15h",
        },
    ];

    let filter: "all" | "issues" = $state("all");

    const issueCount = $derived(
        DUMMY_FEEDS.filter((f) => f.status === "error").length,
    );

    const visible = $derived(
        DUMMY_FEEDS.filter((f) => filter === "all" || f.status === "error")
            // errors first, then pending, then healthy — surface what
            // needs attention before the rest.
            .toSorted((a, b) => rank(a) - rank(b)),
    );

    function rank(f: Feed): number {
        return f.status === "error" ? 0 : f.status === "pending" ? 1 : 2;
    }

    function postsLabel(n: number): string {
        if (n === 0) return "no new posts";
        if (n === 1) return "1 new post";
        return `${n} new posts`;
    }
</script>

<div class="toolbar">
    <div class="filter" role="tablist" aria-label="filter feeds">
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
            aria-selected={filter === "issues"}
            onclick={() => (filter = "issues")}
        >
            needs attention{issueCount > 0 ? ` (${issueCount})` : ""}
        </button>
    </div>
</div>
<hr class="rule-heavy" />

<div class="wire-head">
    <h2 class="section-label">the wire</h2>
    <span class="section-label count" class:has-issue={issueCount > 0}>
        {DUMMY_FEEDS.length} feeds · {issueCount > 0
            ? `${issueCount} need${issueCount === 1 ? "s" : ""} attention`
            : "all healthy"}
    </span>
</div>

{#if visible.length === 0}
    <div class="empty">
        <span class="display">All feeds are healthy.</span>
        <span class="section-label">nothing needs your attention</span>
    </div>
{:else}
    <ul class="feeds">
        {#each visible as feed (feed.id)}
            <li class="row" class:has-issue={feed.status === "error"}>
                <span class="dot {feed.status}"></span>
                <div class="row-text">
                    <strong class="display feed-name">{feed.name}</strong>
                    {#if feed.status === "ok"}
                        <span class="row-meta section-label">
                            checked {feed.checkedAgo} ·
                            {postsLabel(feed.newPosts ?? 0)} · next in
                            {feed.nextIn}
                        </span>
                    {:else if feed.status === "error"}
                        <span
                            class="row-meta section-label error-meta"
                            title={feed.errorDetail}
                        >
                            {#if feed.errorKind === "notFound"}
                                checked {feed.checkedAgo} · feed not found ·
                                check the url, or remove this feed
                            {:else}
                                checked {feed.checkedAgo} · temporarily
                                unreachable · retrying in {feed.nextIn}
                            {/if}
                        </span>
                    {:else}
                        <span class="row-meta section-label">
                            added {feed.addedAgo} · queued for first check
                        </span>
                    {/if}
                </div>
            </li>
        {/each}
    </ul>
{/if}

<style>
    .toolbar {
        display: flex;
        justify-content: space-between;
        align-items: baseline;
        gap: 1rem;
        flex-wrap: wrap;
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

    .wire-head {
        display: flex;
        justify-content: space-between;
        align-items: baseline;
        gap: 1rem;
        margin: 1.5rem 0 0.4rem;
    }

    .wire-head h2 {
        margin: 0;
    }

    .count.has-issue {
        color: var(--accent);
    }

    .feeds {
        list-style: none;
        margin: 0;
        padding: 0;
    }

    .row {
        display: flex;
        align-items: flex-start;
        gap: 0.75rem;
        padding: 1rem 0;
        border-bottom: 1px solid var(--rule);
    }

    /* status dot reuses the read/unread vocabulary from the posts page:
       quiet hollow ring = fine, filled accent = needs your eyes */
    .dot {
        flex-shrink: 0;
        width: 8px;
        height: 8px;
        border-radius: 50%;
        margin-top: 0.45rem;
    }

    .dot.ok {
        background: transparent;
        border: 1px solid var(--rule);
    }

    .dot.error {
        background: var(--accent);
    }

    .dot.pending {
        background: transparent;
        border: 1px dashed var(--ink-faint);
    }

    .row-text {
        display: flex;
        flex-direction: column;
        gap: 0.3rem;
        min-width: 0;
    }

    .feed-name {
        font-size: 1.15rem;
        line-height: 1.2;
        font-variation-settings:
            "opsz" 40,
            "SOFT" 40,
            "WONK" 0;
    }

    .row.has-issue .feed-name {
        color: var(--accent);
    }

    .row-meta {
        letter-spacing: 0.08em;
    }

    .error-meta {
        color: var(--accent);
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
        .link-tab {
            transition: none;
        }
    }
</style>
