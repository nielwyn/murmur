<script lang="ts">
    import { api, type User } from "./lib/api";
    import AuthForm from "./lib/AuthForm.svelte";
    import Feeds from "./lib/Feeds.svelte";
    import Posts from "./lib/Posts.svelte";
    import Status from "./lib/Status.svelte";

    const VIEWS = ["posts", "feeds", "status"] as const;
    type View = (typeof VIEWS)[number];
    const VIEW_KEY = "murmur-view";

    function loadView(): View {
        const saved = localStorage.getItem(VIEW_KEY);
        return saved && (VIEWS as readonly string[]).includes(saved)
            ? (saved as View)
            : "posts";
    }

    let user: User | null = $state(null);
    let checking = $state(true);
    let view: View = $state(loadView());

    $effect(() => {
        localStorage.setItem(VIEW_KEY, view);
    });

    const dateline = new Date()
        .toLocaleDateString("en-US", {
            weekday: "long",
            year: "numeric",
            month: "long",
            day: "numeric",
        })
        .toUpperCase();

    // Restore the session from the auth cookie, if there is one.
    api.me()
        .then((u) => (user = u))
        .catch(() => (user = null))
        .finally(() => (checking = false));

    async function logout() {
        await api.logout();
        user = null;
    }
</script>

<main class="container">
    {#if checking}
        <p aria-busy="true" class="center">loading…</p>
    {:else if user === null}
        <AuthForm onAuthed={(u) => (user = u)} />
    {:else}
        <header class="masthead">
            <div class="dateline">
                <span class="section-label">{dateline}</span>
                <span class="section-label">
                    {user.username} ·
                    <button class="link-button" onclick={logout}>
                        log out
                    </button>
                </span>
            </div>
            <hr class="rule-heavy" />
            <h1 class="display">murmur</h1>
            <hr class="rule" />
            <nav class="paper-nav section-label" aria-label="sections">
                <button
                    class="nav-tab"
                    aria-current={view === "posts" ? "page" : undefined}
                    onclick={() => (view = "posts")}
                >
                    posts
                </button>
                <button
                    class="nav-tab"
                    aria-current={view === "feeds" ? "page" : undefined}
                    onclick={() => (view = "feeds")}
                >
                    feeds
                </button>
                <button
                    class="nav-tab"
                    aria-current={view === "status" ? "page" : undefined}
                    onclick={() => (view = "status")}
                >
                    status
                </button>
            </nav>
            <hr class="rule" />
        </header>
        {#if view === "posts"}
            <Posts />
        {:else if view === "feeds"}
            <Feeds />
        {:else}
            <Status />
        {/if}
    {/if}
</main>

<style>
    main {
        max-width: 42rem;
        padding-block: 2rem;
    }

    .center {
        text-align: center;
        margin-top: 4rem;
    }

    .masthead {
        margin-bottom: 2.5rem;
    }

    /* section nav sits between two rules, like a paper's index bar */
    .paper-nav {
        display: flex;
        justify-content: flex-start;
        gap: 1.5rem;
        padding: 0.55rem 0;
    }

    .nav-tab {
        all: unset;
        cursor: pointer;
        font: inherit;
        letter-spacing: inherit;
        text-transform: inherit;
        color: inherit;
        padding-bottom: 2px;
        border-bottom: 1px solid transparent;
    }

    .nav-tab:hover {
        color: var(--accent);
    }

    .nav-tab[aria-current="page"] {
        color: var(--ink);
        border-bottom-color: var(--accent);
    }

    .nav-tab:focus-visible {
        outline: 2px solid var(--accent);
        outline-offset: 3px;
    }

    .dateline {
        display: flex;
        justify-content: space-between;
        align-items: baseline;
        padding-bottom: 0.5rem;
    }

    .masthead h1 {
        font-size: clamp(3rem, 10vw, 4.5rem);
        line-height: 1.05;
        margin: 0;
        padding: 0.5rem 0 0.75rem;
    }

    .link-button {
        all: unset;
        cursor: pointer;
        font: inherit;
        letter-spacing: inherit;
        text-transform: inherit;
        color: inherit;
        border-bottom: 1px solid transparent;
    }

    .link-button:hover {
        color: var(--accent);
        border-bottom-color: currentColor;
    }
</style>
