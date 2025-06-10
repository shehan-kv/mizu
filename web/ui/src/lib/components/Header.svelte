<script lang="ts">
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import { toggleTheme } from '$lib/utils/theme';
	import BellIcon from './icons/BellIcon.svelte';
	import MenuIcon from './icons/MenuIcon.svelte';
	import MoonIcon from './icons/MoonIcon.svelte';
	import SignOutIcon from './icons/SignOutIcon.svelte';
	import SunIcon from './icons/SunIcon.svelte';
	import UserGearIcon from './icons/UserGearIcon.svelte';
	import UserIcon from './icons/UserIcon.svelte';

	let { openMobileMenu }: { openMobileMenu: () => void } = $props();
</script>

<header class="flex items-center justify-between border-b px-8 py-2">
	<div class="flex gap-4">
		<button class="cursor-pointer lg:hidden" onclick={openMobileMenu}>
			<MenuIcon class="size-5" />
		</button>
		<img src="/logo-light.svg" alt="MizuPM logo" class="w-15 hidden dark:block" />
		<img src="/logo-dark.svg" alt="MizuPM logo" class="w-15 block dark:hidden" />
	</div>

	<div class="flex items-center gap-6 text-neutral-700 dark:text-neutral-300">
		<button
			class="cursor-pointer p-1 hover:text-neutral-950 dark:hover:text-neutral-50"
			onclick={toggleTheme}
		>
			<MoonIcon class="block size-5 dark:hidden" />
			<SunIcon class="hidden size-5 dark:block" />
		</button>

		<Popover.Root>
			<Popover.Trigger class="cursor-pointer p-1 hover:text-neutral-950 dark:hover:text-neutral-50">
				<BellIcon class="size-5" />
			</Popover.Trigger>
			<Popover.Content class="mr-4 mt-2 text-sm">No Notifications Yet</Popover.Content>
		</Popover.Root>

		<DropdownMenu.Root>
			<DropdownMenu.Trigger
				class="cursor-pointer p-1 hover:text-neutral-950 dark:hover:text-neutral-50"
			>
				<UserIcon class="size-5" />
			</DropdownMenu.Trigger>
			<DropdownMenu.Content class="mr-4 mt-2">
				<DropdownMenu.Item class="py-2">
					<span class="flex items-center gap-3"><SignOutIcon class="size-5" />Sign Out</span>
				</DropdownMenu.Item>
				<DropdownMenu.Item class="py-2">
					<span class="flex items-center gap-3">
						<UserGearIcon class="size-5" />Profile Settings
					</span>
				</DropdownMenu.Item>
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</div>
</header>
