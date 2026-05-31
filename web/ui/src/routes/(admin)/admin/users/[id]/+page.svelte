<script lang="ts">
	import { page } from '$app/state';
	import * as Dialog from '$lib/components/dialogs';
	import { createDialogState } from '$lib/components/dialogs/createDialogState.svelte';
	import { formatDate } from '$lib/utils/formatDate';
	import Checks from 'phosphor-svelte/lib/Checks';
	import Info from 'phosphor-svelte/lib/Info';

	let id = page.params.id || '';

	const statusDialog = createDialogState();
	const verifyEmailDialog = createDialogState();
	const manageProjectsDialog = createDialogState();
	const editDialog = createDialogState();
	const deleteDialog = createDialogState();

	type StatusAction = 'activate' | 'deactivate';
	let statusAction: StatusAction | null = $state(null);
	function openStatusDialog(action: StatusAction) {
		statusAction = action;
		statusDialog.open();
	}
</script>

<svelte:head>
	<title>View User</title>
</svelte:head>

<div>
	<!-- Controls -->
	<div>
		<ul class="ml-auto flex w-fit gap-1 text-xs">
			<button
				onclick={() => openStatusDialog('activate')}
				class="cursor-pointer rounded bg-neutral-200 px-4 py-2
						text-xs transition hover:bg-neutral-300 dark:bg-neutral-900
						dark:hover:bg-neutral-800"
			>
				Activate User
			</button>
			<button
				onclick={() => openStatusDialog('deactivate')}
				class="cursor-pointer rounded bg-neutral-200 px-4 py-2
						text-xs transition hover:bg-neutral-300 dark:bg-neutral-900
						dark:hover:bg-neutral-800"
			>
				Deactivate User
			</button>
			<button
				onclick={verifyEmailDialog.open}
				class="cursor-pointer rounded bg-neutral-200 px-4 py-2
						text-xs transition hover:bg-neutral-300 dark:bg-neutral-900
						dark:hover:bg-neutral-800"
			>
				Resend Verification email
			</button>
			<button
				onclick={manageProjectsDialog.open}
				class="cursor-pointer rounded bg-neutral-200 px-4 py-2
						text-xs transition hover:bg-neutral-300 dark:bg-neutral-900
						dark:hover:bg-neutral-800"
			>
				Manage Projects
			</button>
			<button
				onclick={editDialog.open}
				class="cursor-pointer rounded bg-neutral-200 px-4 py-2
						text-xs transition hover:bg-neutral-300 dark:bg-neutral-900
						dark:hover:bg-neutral-800"
			>
				Edit User
			</button>
			<button
				onclick={deleteDialog.open}
				class="cursor-pointer rounded bg-red-300 px-4 py-2 text-xs text-red-950 transition
						hover:bg-red-400 dark:bg-red-900 dark:text-red-50
						dark:hover:bg-red-800"
			>
				Delete
			</button>
		</ul>
	</div>

	<!-- User details section -->
	<div class="max-w-xl space-y-4 rounded-lg bg-neutral-900 p-4">
		<div class="flex items-center gap-4">
			<span
				class="flex size-32 items-center justify-center overflow-hidden
				rounded-full bg-neutral-800 text-4xl text-neutral-600"
			>
				J
			</span>
			<div class="space-y-2.5">
				<div>
					<div class="flex items-center gap-2">
						<p class="text-lg">Jane Doe</p>
					</div>
					<p class="flex gap-2 text-sm">
						<span> Full Stack Developer </span>
						<span> | </span>
						<span> Staff </span>
					</p>
				</div>

				<div class="space-y-1">
					<div class="flex gap-4">
						<div>
							<p class="text-xs text-neutral-400">Last Sign-in</p>
							<p class="text-xs">{formatDate(new Date())}</p>
						</div>
						<div>
							<p class="text-xs text-neutral-400">Created On</p>
							<p class="text-xs">{formatDate(new Date())}</p>
						</div>
					</div>
					<div class="flex gap-1">
						<div
							class="flex w-fit items-center gap-1 rounded
							bg-green-950 px-2 py-1 text-xs text-green-300"
						>
							<Checks />
							<p>Active</p>
						</div>
						<div
							class="flex w-fit items-center gap-1 rounded
							bg-neutral-800 px-2 py-1 text-xs text-neutral-300"
						>
							<Info weight="fill" />
							<p>Pending Verification</p>
						</div>
					</div>
				</div>
			</div>
		</div>
		<div class="max-h-50 overflow-scroll">
			<p class="text-sm tracking-wide">
				A curious and quietly ambitious problem-solver with a background in environmental science
				and a growing interest in product design. Based in a mid-sized coastal city, Alex spends
				weekdays analyzing data and weekends exploring local hiking trails, experimenting with new
				recipes, or tinkering with side projects that blend creativity and technology.
			</p>
		</div>
	</div>

	<!-- Rest of the info here -->
</div>

<br />

<p>List of projects assigned to user</p>
<p>List of invoices for user</p>
<p>List of ch reqs for user</p>
<p>List of tasks assigned to user</p>
<p>Basically everything else for this user</p>

{#if statusAction}
	<Dialog.UserStatusConfirm bind:open={statusDialog.isOpen} userId={id} status={statusAction} />
{/if}

<Dialog.ResendVerifyEmailConfirm bind:open={verifyEmailDialog.isOpen} userId={id} />
<Dialog.UserDeleteConfirm bind:open={deleteDialog.isOpen} userId={id} />
<Dialog.ManageUserProjects bind:open={manageProjectsDialog.isOpen} userId={id} />
