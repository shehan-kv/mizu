<script lang="ts">
	import Moon from 'phosphor-svelte/lib/Moon';
	import Sun from 'phosphor-svelte/lib/Sun';
	import Bell from 'phosphor-svelte/lib/Bell';
	import User from 'phosphor-svelte/lib/User';
	import SignOut from 'phosphor-svelte/lib/SignOut';
	import UserGear from 'phosphor-svelte/lib/UserGear';
	import List from 'phosphor-svelte/lib/List';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import { toggleTheme } from '$lib/utils/theme';
	import { signOut } from '$lib/api/users';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { ApiError } from '$lib/api/client';
	import { toast } from 'svelte-sonner';
	import CircleNotch from 'phosphor-svelte/lib/CircleNotch';

	let { openMobileMenu }: { openMobileMenu: () => void } = $props();

	let signoutLoading = $state(false);

	function handleSignOut() {
		try {
			signoutLoading = true;
			signOut();
			goto(resolve('/sign-in'));
		} catch (error) {
			if (error instanceof ApiError) {
				toast.error(error.message);
			} else {
				toast.error('Could Not Sign Out');
			}
		}
	}
</script>

<header
	class="flex items-center justify-between rounded bg-neutral-50 px-8 py-2 dark:bg-neutral-950"
>
	<div class="flex gap-4">
		<button class="cursor-pointer lg:hidden" onclick={openMobileMenu}>
			<List />
		</button>
		<img src="/assets/logo-light.svg" alt="MizuPM logo" class="hidden w-15 dark:block" />
		<img src="/assets/logo-dark.svg" alt="MizuPM logo" class="block w-15 dark:hidden" />
	</div>

	<div class="flex items-center gap-5 text-neutral-700 dark:text-neutral-300">
		<button
			class="cursor-pointer p-1 hover:text-neutral-950 dark:hover:text-neutral-50"
			onclick={toggleTheme}
		>
			<Moon size={20} class="block dark:hidden" />
			<Sun size={20} class="hidden dark:block" />
		</button>

		<Popover.Root>
			<Popover.Trigger class="cursor-pointer p-1 hover:text-neutral-950 dark:hover:text-neutral-50">
				<Bell size={20} />
			</Popover.Trigger>
			<Popover.Content class="mt-2 mr-4 text-sm">No Notifications Yet</Popover.Content>
		</Popover.Root>

		<DropdownMenu.Root>
			<DropdownMenu.Trigger
				class="cursor-pointer p-1 hover:text-neutral-950 dark:hover:text-neutral-50"
			>
				<User size={20} />
			</DropdownMenu.Trigger>
			<DropdownMenu.Content class="mt-2 mr-4">
				<DropdownMenu.Item class="py-2" onclick={handleSignOut} disabled={signoutLoading}>
					<button class="flex items-center gap-3">
						{#if signoutLoading}
							<CircleNotch />
						{:else}
							<SignOut />
						{/if}
						Sign Out
					</button>
				</DropdownMenu.Item>
				<DropdownMenu.Item class="py-2">
					<span class="flex items-center gap-3">
						<UserGear />Profile Settings
					</span>
				</DropdownMenu.Item>
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</div>
</header>
