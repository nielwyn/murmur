<script lang="ts">
    import { api, ApiError, type Feed } from "./api";

    let feeds: Feed[] = $state([]);
    let following: Set<string> = $state(new Set());
    let loading = $state(true);
    let error = $state("");

    let link = $state("");
    let adding = $state(false);

    async function load() {
        loading = true;
        error = "";
        try {
            const [allFeeds, follows] = await Promise.all([
                api.listFeeds(),
                api.listFollowing(),
            ]);
            feeds = allFeeds;
            following = new Set(follows.map((f) => f.feed_id));
        } catch (e) {
            error = e instanceof ApiError ? e.message : "could not load feeds";
        } finally {
            loading = false;
        }
    }

    async function addFeed(event: SubmitEvent) {
        event.preventDefault();
        error = "";
        adding = true;
        try {
            // Create the feed, then follow it — two separate endpoints.
            const feed = await api.createFeed(link);
            await api.followFeed(feed.id);
            link = "";
            await load();
        } catch (e) {
            error = e instanceof ApiError ? e.message : "could not add feed";
        } finally {
            adding = false;
        }
    }

    async function toggleFollow(feed: Feed) {
        error = "";
        try {
            if (following.has(feed.id)) {
                await api.unfollowFeed(feed.id);
                following.delete(feed.id);
            } else {
                await api.followFeed(feed.id);
                following.add(feed.id);
            }
            following = new Set(following);
        } catch (e) {
            error =
                e instanceof ApiError ? e.message : "could not update follow";
        }
    }

    load();
</script>

<section>
    <h2 class="section-label">add a feed</h2>
    <form onsubmit={addFeed}>
        <!-- svelte-ignore a11y_no_redundant_roles -- Pico CSS styles the
             group via the [role=group] selector, not the fieldset itself. -->
        <fieldset role="group">
            <input
                type="url"
                placeholder="https://example.com/rss"
                bind:value={link}
                required
            />
            <button type="submit" aria-busy={adding}>Add</button>
        </fieldset>
    </form>
</section>

{#if error}
    <p class="error">{error}</p>
{/if}

<section>
    <div class="index-head">
        <h2 class="section-label">the index</h2>
        {#if !loading}
            <span class="section-label">
                {feeds.length}
                {feeds.length === 1 ? "feed" : "feeds"} ·
                {following.size} followed
            </span>
        {/if}
    </div>
    <hr class="rule-heavy" />

    {#if loading}
        <p aria-busy="true" class="state">loading…</p>
    {:else if feeds.length === 0}
        <p class="muted state">No feeds yet — add the first one above.</p>
    {:else}
        <ul>
            {#each feeds as feed (feed.id)}
                <li>
                    <div class="feed-info">
                        <strong class="display feed-name">{feed.title}</strong>
                        <span class="feed-meta section-label">
                            <a href={feed.link} target="_blank" rel="noreferrer">
                                {feed.link}
                            </a>
                            {#if feed.creator_name}
                                · added by {feed.creator_name}
                            {/if}
                        </span>
                    </div>
                    <button
                        class={following.has(feed.id)
                            ? "follow secondary outline"
                            : "follow"}
                        onclick={() => toggleFollow(feed)}
                    >
                        {following.has(feed.id) ? "Following" : "Follow"}
                    </button>
                </li>
            {/each}
        </ul>
    {/if}
</section>

<style>
    section {
        margin-bottom: 2.5rem;
    }

    h2 {
        margin-bottom: 0.75rem;
    }

    form,
    fieldset {
        margin: 0;
    }

    .index-head {
        display: flex;
        justify-content: space-between;
        align-items: baseline;
        margin-bottom: 0.5rem;
    }

    .index-head h2 {
        margin: 0;
    }

    .state {
        padding-top: 1rem;
    }

    ul {
        list-style: none;
        padding: 0;
        margin: 0;
    }

    li {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 1.5rem;
        padding: 1.1rem 0;
        border-bottom: 1px solid var(--rule);
    }

    .feed-info {
        display: flex;
        flex-direction: column;
        gap: 0.3rem;
        min-width: 0;
    }

    .feed-name {
        font-size: 1.35rem;
        line-height: 1.2;
        font-variation-settings:
            "opsz" 40,
            "SOFT" 40,
            "WONK" 0;
    }

    .feed-meta {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        letter-spacing: 0.08em;
    }

    .feed-meta a {
        color: inherit;
    }

    button.follow {
        margin: 0;
        width: 7.5rem;
        flex-shrink: 0;
        padding: 0.45rem 0;
        font-family: var(--font-mono);
        font-size: 0.7rem;
        text-transform: uppercase;
        letter-spacing: 0.15em;
    }

    /* "Following" reads as a quiet state, not a call to action */
    button.follow.outline {
        color: var(--ink-faint);
        border-color: var(--rule);
    }

    button.follow.outline:hover {
        color: var(--accent);
        border-color: var(--accent);
    }
</style>
