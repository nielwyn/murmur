<script lang="ts">
    import { api, ApiError, type User } from "./api";

    let { onAuthed }: { onAuthed: (user: User) => void } = $props();

    let mode: "login" | "register" = $state("login");
    let username = $state("");
    let email = $state("");
    let password = $state("");
    let error = $state("");
    let busy = $state(false);

    async function submit(event: SubmitEvent) {
        event.preventDefault();
        error = "";
        busy = true;
        try {
            const user =
                mode === "login"
                    ? await api.login(username, password)
                    : await api.register(username, email, password);
            onAuthed(user);
        } catch (e) {
            error = e instanceof ApiError ? e.message : "something went wrong";
        } finally {
            busy = false;
        }
    }
</script>

<div class="cover">
    <hgroup>
        <span class="section-label">no algorithm · your feeds, in order</span>
        <h1 class="display">murmur</h1>
    </hgroup>

    <hr class="rule-heavy" />

    <div class="tabs section-label" role="tablist">
        <button
            role="tab"
            aria-selected={mode === "login"}
            class="link-tab"
            onclick={() => (mode = "login")}
        >
            log in
        </button>
        <span aria-hidden="true">/</span>
        <button
            role="tab"
            aria-selected={mode === "register"}
            class="link-tab"
            onclick={() => (mode = "register")}
        >
            register
        </button>
    </div>

    <form onsubmit={submit}>
        <input placeholder="username" bind:value={username} required />
        {#if mode === "register"}
            <input
                type="email"
                placeholder="email"
                bind:value={email}
                required
            />
        {/if}
        <input
            type="password"
            placeholder="password"
            bind:value={password}
            required
        />

        {#if error}
            <p class="error">{error}</p>
        {/if}

        <button type="submit" aria-busy={busy}>
            {mode === "login" ? "Log in" : "Create account"}
        </button>
    </form>
</div>

<style>
    .cover {
        max-width: 24rem;
        margin: 14vh auto 0;
    }

    hgroup {
        text-align: center;
        margin-bottom: 1rem;
    }

    hgroup h1 {
        font-size: clamp(3.5rem, 14vw, 5rem);
        line-height: 1;
        margin: 0.25rem 0 1.5rem;
    }

    .tabs {
        display: flex;
        justify-content: center;
        gap: 0.75rem;
        margin: 1.25rem 0;
    }

    .link-tab {
        all: unset;
        cursor: pointer;
        font: inherit;
        letter-spacing: inherit;
        text-transform: inherit;
        color: inherit;
        padding-bottom: 2px;
        border-bottom: 1px solid transparent;
    }

    .link-tab:hover {
        color: var(--accent);
    }

    .link-tab[aria-selected="true"] {
        color: var(--ink);
        border-bottom-color: var(--accent);
    }

    form {
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
    }
</style>
