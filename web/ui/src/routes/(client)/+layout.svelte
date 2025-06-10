<script>
	import { page } from '$app/state';
	import Header from '$lib/components/Header.svelte';
	import ChatsIcon from '$lib/components/icons/ChatsIcon.svelte';
	import DashboardIcon from '$lib/components/icons/DashboardIcon.svelte';
	import FolderIcon from '$lib/components/icons/FolderIcon.svelte';
	import InvoiceIcon from '$lib/components/icons/InvoiceIcon.svelte';
	import TicketIcon from '$lib/components/icons/TicketIcon.svelte';
	import UserGearIcon from '$lib/components/icons/UserGearIcon.svelte';
	import XIcon from '$lib/components/icons/XIcon.svelte';
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
					class="block flex items-center gap-2 border-l py-2 pl-4 hover:border-sky-500"
					class:border-sky-500={page.url.pathname == '/'}
					onclick={closeMobileMenu}
				>
					<DashboardIcon class="size-5" /> Dashboard
				</a>
			</li>
			<li>
				<a
					href="/projects"
					class="block flex items-center gap-2 border-l py-2 pl-4 hover:border-sky-500"
					class:border-sky-500={page.url.pathname.startsWith('/projects')}
					onclick={closeMobileMenu}
				>
					<FolderIcon class="size-5" /> Projects
				</a>
			</li>
			<li>
				<a
					href="/messages"
					class="block flex items-center gap-2 border-l py-2 pl-4 hover:border-sky-500"
					class:border-sky-500={page.url.pathname.startsWith('/messages')}
					onclick={closeMobileMenu}
				>
					<ChatsIcon class="size-5" /> Messages
				</a>
			</li>
			<li>
				<a
					href="/invoices-and-quotes"
					class="block flex items-center gap-2 border-l py-2 pl-4 hover:border-sky-500"
					class:border-sky-500={page.url.pathname.startsWith('/invoices-and-quotes')}
					onclick={closeMobileMenu}
				>
					<InvoiceIcon class="size-5" /> Invoices & Quotes
				</a>
			</li>
			<li>
				<a
					href="/support-tickets"
					class="block flex items-center gap-2 border-l py-2 pl-4 hover:border-sky-500"
					class:border-sky-500={page.url.pathname.startsWith('/support-tickets')}
					onclick={closeMobileMenu}
				>
					<TicketIcon class="size-5" /> Support Tickets
				</a>
			</li>
			<li class="mt-auto">
				<button
					class="block flex cursor-pointer items-center gap-2 border-l py-2 pl-4 hover:border-sky-500"
				>
					<UserGearIcon class="size-5" /> Profile Settings
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
					<XIcon class="size-4" /></button
				>
			</div>
			{@render nav()}
		</div>
	</div>
{/if}
