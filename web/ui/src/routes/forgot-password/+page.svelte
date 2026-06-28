<script lang="ts">
	import { ApiError } from '$lib/api/client';
	import { createRecovery } from '$lib/api/users';
	import InputLabel from '$lib/components/InputLabel.svelte';
	import { toggleTheme } from '$lib/utils/theme';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import CircleNotch from 'phosphor-svelte/lib/CircleNotch';
	import Moon from 'phosphor-svelte/lib/Moon';
	import Sun from 'phosphor-svelte/lib/Sun';
	import { toast } from 'svelte-sonner';

	let isLoading = $state(false);
	let isSuccess = $state(false);
	let email = $state('');

	let abort: AbortController | null = null;
	async function onSubmit(e: SubmitEvent) {
		e.preventDefault();

		abort?.abort();
		abort = new AbortController();

		isLoading = true;
		isSuccess = false;

		try {
			await createRecovery({ email }, abort.signal);
			isSuccess = true;
		} catch (err) {
			if (err instanceof ApiError) {
				toast.error(toTitleCase(err.message));
			} else {
				toast.error('An Error Occurred');
			}
			isSuccess = false;
		} finally {
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>Forgot Password</title>
</svelte:head>

<div class="grid h-dvh auto-rows-[min-content_1fr]">
	<div class="p-4 text-right">
		<button
			onclick={toggleTheme}
			class="cursor-pointer rounded-full bg-neutral-950 p-2 text-neutral-50 hover:bg-neutral-200
				hover:text-neutral-950 dark:bg-neutral-50 dark:text-neutral-950 dark:hover:bg-neutral-800
				dark:hover:text-neutral-50"
		>
			<Moon weight="fill" size={20} class="block dark:hidden" />
			<Sun weight="fill" size={20} class="hidden dark:block" />
		</button>
	</div>

	<div class="flex items-center justify-center overflow-hidden p-4">
		<div class="w-full max-w-md p-4">
			<div class="text-neutral-950 dark:text-neutral-50">
				<img src="/assets/logo-light.svg" alt="MizuPM logo" class="hidden dark:block" />
				<img src="/assets/logo-dark.svg" alt="MizuPM logo" class="block dark:hidden" />
			</div>

			<div class="mt-4">
				<p class="text-2xl font-bold">Recover Your Account</p>
				<p>Enter Your Email To Proceed</p>
			</div>

			<form class="mt-10 text-sm" onsubmit={onSubmit}>
				<div class="space-y-4">
					<div class="space-y-1 *:block">
						<InputLabel htmlFor="email" text="Email" required />
						<input
							type="email"
							name="email"
							id="email"
							required
							bind:value={email}
							class="w-full rounded border border-neutral-200 bg-neutral-100
                            p-2 dark:border-neutral-800 dark:bg-neutral-900"
						/>
					</div>
				</div>

				<div class="mt-8">
					{#if isSuccess}
						<p class="rounded bg-green-400 p-4 text-center text-green-950">
							You Will Receive A Password Recovery Link Shortly
						</p>
					{:else}
						<button
							type="submit"
							class="w-full cursor-pointer rounded bg-neutral-950
					        py-3 text-neutral-50 transition hover:bg-neutral-200
					        hover:text-neutral-950 disabled:cursor-default dark:bg-neutral-50 dark:text-neutral-950
					        dark:hover:bg-neutral-800 dark:hover:text-neutral-50"
						>
							{#if isLoading}
								<span class="flex items-center justify-center gap-2">
									<CircleNotch class="animate-spin" /> Requesting
								</span>
							{:else}
								Continue
							{/if}
						</button>
					{/if}
				</div>
			</form>
		</div>
	</div>
</div>
