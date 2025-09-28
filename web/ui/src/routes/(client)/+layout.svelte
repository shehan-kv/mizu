<script>
	import Folder from 'phosphor-svelte/lib/Folder';
	import Chats from 'phosphor-svelte/lib/Chats';
	import Invoice from 'phosphor-svelte/lib/Invoice';
	import UserGear from 'phosphor-svelte/lib/UserGear';
	import Swap from 'phosphor-svelte/lib/Swap';
	import X from 'phosphor-svelte/lib/X';
	import LayoutDashboard from '@lucide/svelte/icons/layout-dashboard';
	import { page } from '$app/state';
	import Header from '$lib/components/Header.svelte';
	import { fade, fly } from 'svelte/transition';

	let { children } = $props();
	let isMobileMenuOpen = $state(false);

	function openMobileMenu() {
		isMobileMenuOpen = true;
	}

	function closeMobileMenu() {
		isMobileMenuOpen = false;
	}
</script>

{#snippet nav()}
	<nav class="h-full">
		<ul class="flex h-full flex-col text-sm">
			<li>
				<a
					href="/"
					class="block flex items-center gap-2 border-l py-2 pl-4
					hover:border-neutral-400 dark:hover:border-neutral-700"
					class:border-sky-500={page.url.pathname == '/'}
					onclick={closeMobileMenu}
				>
					<LayoutDashboard size={20} strokeWidth={1.5} /> Dashboard
				</a>
			</li>
			<li>
				<a
					href="/projects"
					class="block flex items-center gap-2 border-l py-2 pl-4
					hover:border-neutral-400 dark:hover:border-neutral-700"
					class:border-sky-500={page.url.pathname.startsWith('/projects')}
					onclick={closeMobileMenu}
				>
					<Folder size={20} /> Projects
				</a>
			</li>
			<li>
				<a
					href="/messages"
					class="block flex items-center gap-2 border-l py-2 pl-4
					hover:border-neutral-400 dark:hover:border-neutral-700"
					class:border-sky-500={page.url.pathname.startsWith('/messages')}
					onclick={closeMobileMenu}
				>
					<Chats size={20} /> Messages
				</a>
			</li>
			<li>
				<a
					href="/invoices-and-quotes"
					class="block flex items-center gap-2 border-l py-2 pl-4
					hover:border-neutral-400 dark:hover:border-neutral-700"
					class:border-sky-500={page.url.pathname.startsWith('/invoices-and-quotes')}
					onclick={closeMobileMenu}
				>
					<Invoice size={20} /> Invoices & Quotes
				</a>
			</li>
			<li>
				<a
					href="/change-requests"
					class="block flex items-center gap-2 border-l py-2 pl-4
					hover:border-neutral-400 dark:hover:border-neutral-700"
					class:border-sky-500={page.url.pathname.startsWith('/change-requests')}
					onclick={closeMobileMenu}
				>
					<Swap size={20} /> Change Requests
				</a>
			</li>
			<li class="mt-auto">
				<button
					class="block flex cursor-pointer items-center gap-2 border-l py-2 pl-4
					hover:border-neutral-400 dark:hover:border-neutral-700"
				>
					<UserGear size={20} /> Profile Settings
				</button>
			</li>
		</ul>
	</nav>
{/snippet}

<div class="grid h-dvh auto-rows-[min-content_1fr] gap-4">
	<Header {openMobileMenu} />
	<div class="mx-8 mb-8 grid grid-cols-1 gap-4 overflow-auto lg:grid-cols-[14rem_1fr]">
		<div class="hidden pt-4 lg:block">
			{@render nav()}
		</div>

		<div class="grow overflow-y-auto py-4 lg:px-4">
			{@render children()}
		</div>
	</div>
</div>

{#if isMobileMenuOpen}
	<div>
		<div
			transition:fade
			onclick={closeMobileMenu}
			role="button"
			tabindex="0"
			aria-label="Close mobile menu"
			onkeydown={(e) => {
				if (e.key === 'Enter' || e.key === ' ') closeMobileMenu();
			}}
			class="backdrop-blur-xs fixed inset-0 bg-neutral-200/30 dark:bg-neutral-800/30"
		></div>
		<div
			transition:fly={{ x: -50 }}
			class="fixed bottom-2 left-2 top-2 grid max-h-dvh w-full max-w-xs auto-rows-[min-content_1fr] overflow-y-scroll rounded bg-neutral-50 p-6 dark:bg-neutral-900"
		>
			<div class="mb-6 flex items-center justify-between">
				<button onclick={closeMobileMenu} class="ml-auto cursor-pointer rounded-full border p-1.5">
					<X size={18} /></button
				>
			</div>
			{@render nav()}
		</div>
	</div>
{/if}
