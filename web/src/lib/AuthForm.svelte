<script lang="ts">
    import { onMount } from "svelte";
    import { api, ApiError, type User } from "./api";
    import GoogleLogo from "./icons/GoogleLogo.svelte";

    let { onAuthed }: { onAuthed: (user: User) => void } = $props();

    let mode: "login" | "register" = $state("login");
    let username = $state("");
    let email = $state("");
    let password = $state("");
    let error = $state("");
    let busy = $state(false);

    let googlePopup: Window | null = null;

    function googleErrorMessage(code: string): string {
        if (code === "google_email_taken") {
            return "An account with this email already exists — log in with your password instead.";
        }
        return "Google sign-in failed. Please try again.";
    }

    function openGooglePopup() {
        error = "";
        const width = 480;
        const height = 620;
        const left = window.screenX + (window.outerWidth - width) / 2;
        const top = window.screenY + (window.outerHeight - height) / 2;
        googlePopup = window.open(
            "/api/auth/google",
            "murmur-google-login",
            `width=${width},height=${height},left=${left},top=${top}`,
        );
    }

    onMount(() => {
        function handleMessage(event: MessageEvent) {
            // Only trust messages from the popup we opened — it navigates
            // through Google's origin and back to our own API's origin
            // during the flow, so checking event.source (a stable window
            // reference) is more robust here than matching event.origin.
            if (event.source !== googlePopup) return;
            if (!event.data || typeof event.data.type !== "string") return;

            if (event.data.type === "google-auth-success") {
                api
                    .me()
                    .then(onAuthed)
                    .catch(() => {
                        error = "Google sign-in failed. Please try again.";
                    });
            } else if (event.data.type === "google-auth-error") {
                error = googleErrorMessage(event.data.code ?? "");
            }
        }
        window.addEventListener("message", handleMessage);
        return () => window.removeEventListener("message", handleMessage);
    });

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

    <hr class="rule" />

    <button type="button" class="google-button" onclick={openGooglePopup}>
        <GoogleLogo />
        <span>Continue with Google</span>
    </button>
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

    .google-button {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 0.65rem;
        width: 100%;
        margin-top: 1rem;
        padding: 0.6rem 1rem;
        border: 1px solid #dadce0;
        border-radius: 4px;
        background: #fff;
        color: #3c4043;
        font-family: "Google Sans", Roboto, Arial, sans-serif;
        font-size: 0.9rem;
        font-weight: 500;
        text-decoration: none;
        cursor: pointer;
        transition:
            background 0.15s ease,
            box-shadow 0.15s ease;
    }

    .google-button:hover {
        background: #f8f9fa;
        box-shadow: 0 1px 2px rgba(0, 0, 0, 0.15);
    }

    .google-button:focus-visible {
        outline: 2px solid #4285f4;
        outline-offset: 2px;
    }
</style>
