<script lang="ts">
	import TrendUp from 'phosphor-svelte/lib/TrendUp';

	import Checkbox from '$lib/components/ui/checkbox/checkbox.svelte';
	import { toggleTheme } from '$lib/utils/theme';
	import Files from 'phosphor-svelte/lib/Files';
	import Chats from 'phosphor-svelte/lib/Chats';
	import Moon from 'phosphor-svelte/lib/Moon';
	import Sun from 'phosphor-svelte/lib/Sun';

	let signInForm = $state({
		email: '',
		password: '',
		remember_me: false
	});
	function signIn() {
		// TODO: Implement sign in functionality
	}
</script>

<svelte:head>
	<title>Sign In</title>
</svelte:head>

<div
	class="grid h-dvh grid-cols-1 p-4 lg:grid-cols-[1fr_28rem] dark:bg-neutral-950 dark:text-neutral-50"
>
	<div
		class="lg: flex hidden size-full items-center justify-center rounded bg-[url('/wave.svg')] px-8 lg:flex dark:bg-[url('/wave-dark.svg')]"
	>
		<div class="@container w-full max-w-3xl rounded bg-neutral-50 p-8 dark:bg-neutral-900">
			<div>
				<span class="inline-block rounded-full bg-neutral-200 p-2 dark:bg-neutral-800">
					<TrendUp size={20} />
				</span>
				<p class="@lg:text-base mt-2 text-sm font-bold">Track Progress</p>
				<p class="max-w-2/3 @lg:text-sm mt-1 text-xs">
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
					<p class="@lg:text-base mt-2 text-sm font-bold">Sign Contracts</p>
					<p class="@lg:text-sm mt-1 mt-1 text-xs">
						Accelerate deal closures with integrated contract signing workflows — no external tools.
						Simply review, sign, and proceed with confidence.
					</p>
				</div>
				<div>
					<span class="inline-block rounded-full bg-neutral-200 p-2 dark:bg-neutral-800">
						<Chats size={20} />
					</span>
					<p class="@lg:text-base mt-2 text-sm font-bold">Unified Chat</p>
					<p class="@lg:text-sm mt-1 mt-1 text-xs">
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
				<img src="/logo-light.svg" alt="MizuPM logo" class="hidden dark:block" />
				<img src="/logo-dark.svg" alt="MizuPM logo" class="block dark:hidden" />
			</div>

			<div class="mt-4">
				<p class="text-2xl font-bold">Welcome Back!</p>
				<p>Please sign in to continue</p>
			</div>

			<form class="mt-10 text-sm" onsubmit={signIn}>
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
					<Checkbox id="remember_me" bind:checked={signInForm.remember_me} />
					<label for="remember_me">Remember Me</label>
				</div>

				<button
					type="submit"
					class="mt-8 w-full cursor-pointer rounded bg-neutral-950
					py-3 text-neutral-50 transition hover:bg-neutral-200
					hover:text-neutral-950 dark:bg-neutral-50 dark:text-neutral-950 dark:hover:bg-neutral-800
					dark:hover:text-neutral-50"
				>
					Continue
				</button>
			</form>

			<a
				href="/"
				class="mt-4 block text-sm text-neutral-700 transition hover:text-neutral-950
			dark:text-neutral-400 dark:hover:text-neutral-50"
			>
				Forgot your password ?
			</a>
		</div>
	</div>
</div>
