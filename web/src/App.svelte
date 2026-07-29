<script lang="ts">
    import { api, type User } from "./lib/api";
    import AuthForm from "./lib/AuthForm.svelte";
    import Feeds from "./lib/Feeds.svelte";
    import Posts from "./lib/Posts.svelte";

    const VIEWS = ["posts", "feeds"] as const;
    type View = (typeof VIEWS)[number];
    const VIEW_KEY = "murmur-view";

    function loadView(): View {
        const saved = localStorage.getItem(VIEW_KEY);
        return saved && (VIEWS as readonly string[]).includes(saved)
            ? (saved as View)
            : "posts";
    }

    type Theme = "light" | "dark";
    const THEME_KEY = "murmur-theme";

    function loadTheme(): Theme {
        const saved = localStorage.getItem(THEME_KEY);
        if (saved === "light" || saved === "dark") return saved;
        return matchMedia("(prefers-color-scheme: dark)").matches
            ? "dark"
            : "light";
    }

    let user: User | null = $state(null);
    let checking = $state(true);
    let view: View = $state(loadView());
    let theme: Theme = $state(loadTheme());

    $effect(() => {
        localStorage.setItem(VIEW_KEY, view);
    });

    // index.html's inline script sets data-theme before first paint for the
    // saved-preference case; this keeps it in sync after that (including the
    // very first explicit choice, when nothing was saved yet).
    $effect(() => {
        document.documentElement.dataset.theme = theme;
        localStorage.setItem(THEME_KEY, theme);
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

    function toggleTheme() {
        theme = theme === "dark" ? "light" : "dark";
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
                <span class="section-label dateline-controls">
                    {dateline}
                    <button
                        class="theme-switch"
                        role="switch"
                        aria-checked={theme === "dark"}
                        aria-label={theme === "dark"
                            ? "switch to light mode"
                            : "switch to dark mode"}
                        onclick={toggleTheme}
                    >
                        <span class="theme-switch-track">
                            <span class="theme-switch-knob">
                                {#if theme === "dark"}
                                    <svg
                                        class="theme-icon"
                                        viewBox="0 0 24 24"
                                        fill="currentColor"
                                        aria-hidden="true"
                                    >
                                        <path
                                            d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"
                                        />
                                    </svg>
                                {:else}
                                    <svg
                                        class="theme-icon"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                        stroke-linecap="round"
                                        aria-hidden="true"
                                    >
                                        <circle cx="12" cy="12" r="4" />
                                        <path
                                            d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41"
                                        />
                                    </svg>
                                {/if}
                            </span>
                        </span>
                    </button>
                </span>
                <span class="section-label dateline-controls">
                    <span class="username" title={user.username}
                        >{user.username}</span
                    >
                    ·
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
            </nav>
            <hr class="rule" />
        </header>
        {#if view === "posts"}
            <Posts onGoToFeeds={() => (view = "feeds")} />
        {:else}
            <Feeds />
        {/if}
    {/if}
</main>

<style>
    main {
        max-width: 42rem;
        padding-block: 2rem;
        /* Pico's .container zeroes this out past 576px, assuming its own
           narrower max-width always leaves an auto-margin gutter — false
           once max-width is overridden wider, so pin it explicitly. */
        padding-inline: var(--pico-spacing);
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

    .dateline-controls {
        display: inline-flex;
        align-items: center;
        gap: 0.5rem;
        min-width: 0;
    }

    .username {
        display: inline-block;
        max-width: 9rem;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        vertical-align: bottom;
    }

    .theme-switch {
        all: unset;
        cursor: pointer;
        display: inline-flex;
        vertical-align: middle;
    }

    .theme-switch-track {
        display: inline-flex;
        align-items: center;
        width: 2.5rem;
        height: 1.35rem;
        padding: 0.15rem;
        border-radius: 999px;
        background: var(--rule);
        transition: background-color 0.25s ease;
    }

    .theme-switch[aria-checked="true"] .theme-switch-track {
        background: var(--ink-faint);
    }

    .theme-switch-knob {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 1.05rem;
        height: 1.05rem;
        border-radius: 50%;
        background: var(--paper-raised);
        color: var(--ink-soft);
        box-shadow: 0 1px 2px rgba(0, 0, 0, 0.25);
        transform: translateX(0);
        transition: transform 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
    }

    .theme-switch[aria-checked="true"] .theme-switch-knob {
        transform: translateX(1.15rem);
        color: var(--accent);
    }

    .theme-icon {
        width: 0.65rem;
        height: 0.65rem;
    }

    .theme-switch:focus-visible .theme-switch-track {
        outline: 2px solid var(--accent);
        outline-offset: 2px;
    }

    @media (prefers-reduced-motion: reduce) {
        .theme-switch-track,
        .theme-switch-knob {
            transition: none;
        }
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
