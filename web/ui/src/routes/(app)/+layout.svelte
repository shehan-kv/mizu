<script lang="ts">
	import Folder from 'phosphor-svelte/lib/Folder';
	import Chats from 'phosphor-svelte/lib/Chats';
	import Invoice from 'phosphor-svelte/lib/Invoice';
	import UserGear from 'phosphor-svelte/lib/UserGear';
	import * as Dialog from '$lib/components/dialogs';
	import X from 'phosphor-svelte/lib/X';
	import LayoutDashboard from '@lucide/svelte/icons/layout-dashboard';
	import { page } from '$app/state';
	import Header from '$lib/components/Header.svelte';
	import { fade, fly } from 'svelte/transition';
	import FileText from 'phosphor-svelte/lib/FileText';
	import User from 'phosphor-svelte/lib/User';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import { messageStore } from '$lib/messages/messageStore.svelte';
	import { channelStore } from '$lib/messages/channelStore.svelte';
	import type { Message } from '$lib/api/messages';
	import { auth } from '$lib/auth/auth.svelte';
	import FullScreenSpinner from '$lib/components/FullScreenSpinner.svelte';
	import FullScreenErrorMessage from '$lib/components/FullScreenErrorMessage.svelte';
	import { createDialogState } from '$lib/components/dialogs/createDialogState.svelte';

	let { children } = $props();
	let isMobileMenuOpen = $state(false);

	function openMobileMenu() {
		isMobileMenuOpen = true;
	}

	function closeMobileMenu() {
		isMobileMenuOpen = false;
	}

	let viewProfile = createDialogState();

	const totalUnread = $derived(
		Object.values(messageStore.state.unreadCounts).reduce((sum, count) => sum + count, 0)
	);

	onMount(() => {
		let sse: EventSource | null = null;

		auth.init().then(() => {
			if (!auth.role) return;

			sse = new EventSource('/api/v1/events');

			sse.addEventListener('integration.messaging.broadcast', (event) => {
				const message = JSON.parse((event as MessageEvent).data) as Message;

				if (message.channelId === messageStore.state.activeChannelId) {
					messageStore.appendMessage(message);
				} else {
					messageStore.incrementUnread(message.channelId);
				}

				channelStore.updateActivity(message.channelId, message.createdAt);
			});
		});

		return () => {
			sse?.close();
		};
	});
</script>

{#snippet nav()}
	<nav class="h-full">
		<ul class="flex h-full flex-col gap-y-0.5 text-sm">
			<li>
				<a
					href={resolve('/')}
					class="flex items-center gap-2 rounded py-2
					pl-4 transition hover:bg-neutral-200 dark:hover:bg-neutral-800"
					class:bg-neutral-200={page.url.pathname == '/'}
					class:dark:bg-neutral-800={page.url.pathname == '/'}
					onclick={closeMobileMenu}
				>
					<LayoutDashboard size={20} strokeWidth={1.5} /> Dashboard
				</a>
			</li>
			<li>
				<a
					href={resolve('/projects')}
					class="flex items-center gap-2 rounded py-2
					pl-4 transition hover:bg-neutral-200 dark:hover:bg-neutral-800"
					class:bg-neutral-200={page.url.pathname.startsWith('/projects')}
					class:dark:bg-neutral-800={page.url.pathname.startsWith('/projects')}
					onclick={closeMobileMenu}
				>
					<Folder size={20} /> Projects
				</a>
			</li>
			<li>
				<a
					href={resolve('/contracts')}
					class="flex items-center gap-2 rounded py-2
					pl-4 transition hover:bg-neutral-200 dark:hover:bg-neutral-800"
					class:bg-neutral-200={page.url.pathname.startsWith('/contracts')}
					class:dark:bg-neutral-800={page.url.pathname.startsWith('/contracts')}
					onclick={closeMobileMenu}
				>
					<FileText size={20} /> Contracts
				</a>
			</li>
			<li>
				<a
					href={resolve('/messages')}
					class="flex items-center gap-2 rounded py-2
					pl-4 transition hover:bg-neutral-200 dark:hover:bg-neutral-800"
					class:bg-neutral-200={page.url.pathname.startsWith('/messages')}
					class:dark:bg-neutral-800={page.url.pathname.startsWith('/messages')}
					onclick={closeMobileMenu}
				>
					<Chats size={20} /> Messages
					{#if totalUnread > 0}
						<span class="inline-block size-1.5 rounded-full bg-red-500"></span>
					{/if}
				</a>
			</li>
			<li>
				<a
					href={resolve('/invoices-and-quotes')}
					class="flex items-center gap-2 rounded py-2
					pl-4 transition hover:bg-neutral-200 dark:hover:bg-neutral-800"
					class:bg-neutral-200={page.url.pathname.startsWith('/invoices-and-quotes')}
					class:dark:bg-neutral-800={page.url.pathname.startsWith('/invoices-and-quotes')}
					onclick={closeMobileMenu}
				>
					<Invoice size={20} /> Invoices & Quotes
				</a>
			</li>
			{#if auth.role == 'administrator'}
				<li>
					<a
						href={resolve('/users')}
						class="flex items-center gap-2 rounded py-2
					pl-4 transition hover:bg-neutral-200 dark:hover:bg-neutral-800"
						class:bg-neutral-200={page.url.pathname.startsWith('/users')}
						class:dark:bg-neutral-800={page.url.pathname.startsWith('/users')}
						onclick={closeMobileMenu}
					>
						<User size={20} /> Users
					</a>
				</li>
			{/if}
			<li class="mt-auto">
				<button
					onclick={viewProfile.open}
					class="flex w-full cursor-pointer items-center gap-2 rounded py-2
					pl-4 transition hover:bg-neutral-200 dark:hover:bg-neutral-800"
				>
					<UserGear size={20} /> Profile Settings
				</button>
			</li>
		</ul>
	</nav>
{/snippet}

{#if auth.loading}
	<FullScreenSpinner />
{:else if auth.error}
	<FullScreenErrorMessage variant="warn" text="There Was An Error, Please Try Refreshing" />
{:else}
	<div
		class="grid h-dvh auto-rows-[min-content_1fr] gap-2 bg-[url('/assets/wave-inverted.svg')] p-2
	dark:bg-[url('/assets/wave-dark.svg')]"
	>
		<Header {openMobileMenu} openProfileSettings={viewProfile.open} />
		<div class="grid grid-cols-1 gap-2 overflow-auto lg:grid-cols-[15rem_1fr]">
			<div class="hidden rounded bg-neutral-50 p-2 lg:block dark:bg-neutral-950">
				{@render nav()}
			</div>

			<div class="grow overflow-y-auto">
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
				class="fixed inset-0 bg-neutral-200/30 backdrop-blur-xs dark:bg-neutral-800/30"
			></div>
			<div
				transition:fly={{ x: -50 }}
				class="fixed top-2 bottom-2 left-2 grid max-h-dvh w-full max-w-xs auto-rows-[min-content_1fr] overflow-y-scroll rounded bg-neutral-50 p-6 dark:bg-neutral-900"
			>
				<div class="mb-6 flex items-center justify-between">
					<button
						onclick={closeMobileMenu}
						class="ml-auto cursor-pointer rounded-full border p-1.5"
					>
						<X size={18} /></button
					>
				</div>
				{@render nav()}
			</div>
		</div>
	{/if}
{/if}

<Dialog.ViewMe bind:open={viewProfile.isOpen} />
