<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { ApiError } from '$lib/api/client';
	import { verifyUser } from '$lib/api/users';
	import FullScreenErrorMessage from '$lib/components/FullScreenErrorMessage.svelte';
	import FullScreenSpinner from '$lib/components/FullScreenSpinner.svelte';
	import InputLabel from '$lib/components/InputLabel.svelte';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import CircleNotch from 'phosphor-svelte/lib/CircleNotch';
	import { toast } from 'svelte-sonner';

	let { data } = $props();

	let id = page.params.id || '';

	let isLoading = $state(false);
	let verifyForm = $state({
		password: '',
		confirmPassword: ''
	});

	let submitAbort: AbortController | null = null;
	function onSubmit(e: SubmitEvent) {
		e.preventDefault();

		submitAbort?.abort();
		submitAbort = new AbortController();

		isLoading = true;
		verifyUser(id, verifyForm, submitAbort.signal)
			.then(() => {
				goto(resolve('/sign-in'));
			})
			.catch((err) => {
				if (err instanceof ApiError) {
					toast.error(toTitleCase(err.message));
				} else {
					toast.error('An Error Occurred');
				}

				isLoading = false;
			});
	}
</script>

<svelte:head>
	<title>Verify Your Account</title>
</svelte:head>

{#await data.validatePromise}
	<FullScreenSpinner />
{:then result}
	{#if result.loadError}
		{#if result.loadError.status == 404}
			<FullScreenErrorMessage variant="warn" text="Invalid Request" />
		{:else}
			<FullScreenErrorMessage variant="warn" text={toTitleCase(result.loadError.message)} />
		{/if}
	{:else}
		<div class="flex h-dvh items-center justify-center overflow-hidden p-4">
			<div class="w-full max-w-md p-4">
				<div class="text-neutral-950 dark:text-neutral-50">
					<img src="/assets/logo-light.svg" alt="MizuPM logo" class="hidden dark:block" />
					<img src="/assets/logo-dark.svg" alt="MizuPM logo" class="block dark:hidden" />
				</div>

				<div class="mt-4">
					<p class="text-2xl font-bold">Verify Your Account</p>
					<p>Assign A Password To Continue</p>
				</div>

				<form class="mt-10 text-sm" onsubmit={onSubmit}>
					<div class="space-y-4">
						<div class="space-y-1 *:block">
							<InputLabel htmlFor="password" text="Password" required />
							<input
								type="password"
								name="password"
								id="password"
								required
								bind:value={verifyForm.password}
								class="w-full rounded border border-neutral-200 bg-neutral-100
							p-2 dark:border-neutral-800 dark:bg-neutral-900"
							/>
						</div>

						<div class="space-y-1 *:block">
							<InputLabel htmlFor="confirmPassword" text="Confirm Password" required />
							<input
								type="password"
								name="confirmPassword"
								id="confirmPassword"
								required
								bind:value={verifyForm.confirmPassword}
								class="w-full rounded border border-neutral-200 bg-neutral-100 p-2
							dark:border-neutral-800 dark:bg-neutral-900"
							/>
						</div>
					</div>

					<button
						disabled={isLoading}
						type="submit"
						class="mt-8 w-full cursor-pointer rounded bg-neutral-950
					py-3 text-neutral-50 transition hover:bg-neutral-200
					hover:text-neutral-950 disabled:cursor-default dark:bg-neutral-50 dark:text-neutral-950
					dark:hover:bg-neutral-800 dark:hover:text-neutral-50"
					>
						{#if isLoading}
							<span class="flex items-center justify-center gap-2">
								<CircleNotch class="animate-spin" /> Verifying
							</span>
						{:else}
							Continue
						{/if}
					</button>
				</form>
			</div>
		</div>
	{/if}
{/await}
