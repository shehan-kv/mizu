<script lang="ts">
	import Files from 'phosphor-svelte/lib/Files';
	import Chats from 'phosphor-svelte/lib/Chats';
	import TrendUp from 'phosphor-svelte/lib/TrendUp';
	import Moon from 'phosphor-svelte/lib/Moon';
	import Sun from 'phosphor-svelte/lib/Sun';
	import { toast } from 'svelte-sonner';

	import Checkbox from '$lib/components/ui/checkbox/checkbox.svelte';
	import { toggleTheme } from '$lib/utils/theme';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { signIn } from '$lib/api/users';

	let isLoading = $state(false);
	let signInForm = $state({
		email: '',
		password: '',
		rememberMe: false
	});

	function onSubmit(e: SubmitEvent) {
		e.preventDefault();

		isLoading = true;
		signIn(signInForm)
			.then(() => {
				goto(resolve('/'));
			})
			.catch(() => {
				toast.error('Sign-in Failed');
				isLoading = false;
			});
	}
</script>

<svelte:head>
	<title>Sign In</title>
</svelte:head>

<div
	class="grid h-dvh grid-cols-1 p-4 lg:grid-cols-[1fr_28rem] dark:bg-neutral-950 dark:text-neutral-50"
>
	<div
		class="hidden size-full items-center justify-center rounded bg-[url('/assets/wave.svg')] px-8 lg:flex dark:bg-[url('/assets/wave-dark.svg')]"
	>
		<div class="@container w-full max-w-3xl rounded bg-neutral-50 p-8 dark:bg-neutral-900">
			<div>
				<span class="inline-block rounded-full bg-neutral-200 p-2 dark:bg-neutral-800">
					<TrendUp size={20} />
				</span>
				<p class="mt-2 text-sm font-bold @lg:text-base">Track Progress</p>
				<p class="mt-1 max-w-2/3 text-xs @lg:text-sm">
					Organize tasks, monitor progress, and maintain focus through a simple, visual workflow
					with Kanban boards
				</p>
			</div>

			<hr class="my-10 border-neutral-200 dark:border-neutral-800" />

			<div class="flex gap-8">
				<div>
					<span class="inline-block rounded-full bg-neutral-200 p-2 dark:bg-neutral-800">
						<Files size={20} />
					</span>
					<p class="mt-2 text-sm font-bold @lg:text-base">Sign Contracts</p>
					<p class="mt-1 text-xs @lg:text-sm">
						Accelerate deal closures with integrated contract signing workflows — no external tools.
						Simply review, sign, and proceed with confidence.
					</p>
				</div>
				<div>
					<span class="inline-block rounded-full bg-neutral-200 p-2 dark:bg-neutral-800">
						<Chats size={20} />
					</span>
					<p class="mt-2 text-sm font-bold @lg:text-base">Unified Chat</p>
					<p class="mt-1 text-xs @lg:text-sm">
						Foster clear and organized collaboration through built-in messaging, ensuring all
						project-related conversations remain accessible and centralized
					</p>
				</div>
			</div>
		</div>
	</div>
	<div class="grid size-full auto-rows-[min-content_1fr] grid-cols-1 px-4 sm:px-10">
		<div class="text-right">
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

		<div class="mx-auto size-full max-w-md content-center">
			<div class="text-neutral-950 dark:text-neutral-50">
				<img src="/assets/logo-light.svg" alt="MizuPM logo" class="hidden dark:block" />
				<img src="/assets/logo-dark.svg" alt="MizuPM logo" class="block dark:hidden" />
			</div>

			<div class="mt-4">
				<p class="text-2xl font-bold">Welcome Back!</p>
				<p>Please sign in to continue</p>
			</div>

			<form class="mt-10 text-sm" onsubmit={onSubmit}>
				<div class="space-y-4">
					<div class="space-y-1">
						<label for="email" class="block">Email</label>
						<input
							type="email"
							name="email"
							id="email"
							required
							bind:value={signInForm.email}
							class="w-full rounded border border-neutral-200 bg-neutral-100
							p-2 dark:border-neutral-800 dark:bg-neutral-900"
						/>
					</div>

					<div class="space-y-1">
						<label for="password" class="block">Password</label>
						<input
							type="password"
							name="password"
							id="password"
							required
							bind:value={signInForm.password}
							class="w-full rounded border border-neutral-200 bg-neutral-100 p-2
							dark:border-neutral-800 dark:bg-neutral-900"
						/>
					</div>
				</div>

				<div class="mt-4 flex items-center gap-2">
					<Checkbox id="remember_me" bind:checked={signInForm.rememberMe} />
					<label for="remember_me">Remember Me</label>
				</div>

				<button
					disabled={isLoading}
					type="submit"
					class="mt-8 w-full cursor-pointer rounded bg-neutral-950
					py-3 text-neutral-50 transition hover:bg-neutral-200
					hover:text-neutral-950 dark:bg-neutral-50 dark:text-neutral-950 dark:hover:bg-neutral-800
					dark:hover:text-neutral-50"
				>
					{isLoading ? 'Signing in...' : 'Continue'}
				</button>
			</form>

			<a
				href={resolve('/forgot-password')}
				class="mt-4 block text-sm text-neutral-700 transition hover:text-neutral-950
			dark:text-neutral-400 dark:hover:text-neutral-50"
			>
				Forgot your password ?
			</a>
		</div>
	</div>
</div>
