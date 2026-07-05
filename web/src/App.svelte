<script lang="ts">
    import { api, type User } from "./lib/api";
    import AuthForm from "./lib/AuthForm.svelte";
    import Feeds from "./lib/Feeds.svelte";

    let user: User | null = $state(null);
    let checking = $state(true);

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
        </header>
        <Feeds />
    {/if}
</main>

<style>
    main {
        max-width: 42rem;
        padding-block: 2rem;
        height: 100vh;
    }

    .center {
        text-align: center;
        margin-top: 4rem;
    }

    .masthead {
        margin-bottom: 2.5rem;
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
