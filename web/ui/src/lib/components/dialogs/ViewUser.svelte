<script lang="ts">
	import X from 'phosphor-svelte/lib/X';
	import { Dialog } from 'bits-ui';
	import { getUser, type User } from '$lib/api/users';
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';
	import Checks from 'phosphor-svelte/lib/Checks';
	import Info from 'phosphor-svelte/lib/Info';
	import { formatDate } from '$lib/utils/formatDate';
	import { createDialogState } from './createDialogState.svelte';
	import UserStatusConfirm from './UserStatusConfirm.svelte';
	import ResendVerifyEmailConfirm from './ResendVerifyEmailConfirm.svelte';
	import UserDeleteConfirm from './UserDeleteConfirm.svelte';
	import ManageUserProjects from './ManageUserProjects.svelte';
	import EditUser from './EditUser.svelte';
	import { ApiError, BASE_URL } from '$lib/api/client';
	import Spinner from '../Spinner.svelte';
	import ErrorMessage from '../ErrorMessage.svelte';

	interface Props {
		userId: string;
		open: boolean;
		refresh?: () => unknown;
	}

	let { open = $bindable(), userId, refresh }: Props = $props();

	let userPromise: Promise<User> | null = $state(null);
	let abort: AbortController | null = null;
	function load() {
		abort?.abort();
		abort = new AbortController();

		userPromise = getUser(userId, abort?.signal);
	}

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

	$effect(() => {
		if (open) {
			load();
		} else {
			abort?.abort();
		}
	});
</script>

<Dialog.Root bind:open>
	<Dialog.Portal>
		<Dialog.Overlay
			class="data-[state=open]:animate-in data-[state=closed]:animate-out 
			data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed 
			inset-0 z-50 bg-neutral-100/80 dark:bg-black/80"
		/>
		<Dialog.Content
			class="bg-background data-[state=open]:animate-in data-[state=closed]:animate-out 
			data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0 
			data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95
			fixed top-1/2 left-1/2 z-50 grid w-full max-w-xl -translate-x-1/2 -translate-y-1/2 auto-rows-[min-content_1fr] gap-4 rounded outline-hidden 
			duration-250"
		>
			<div class="text-right">
				<Dialog.Close
					class="cursor-pointer rounded-bl bg-neutral-50 px-4 py-2
					transition duration-150
					hover:bg-neutral-950 hover:text-neutral-50 dark:bg-neutral-900
					hover:dark:bg-neutral-50 hover:dark:text-neutral-950"
				>
					<X class="size-3" />
				</Dialog.Close>
			</div>

			<div class="px-6 pb-6">
				{#await userPromise}
					<Spinner />
				{:then user}
					{#if user}
						<div class="space-y-4">
							<div class="mx-auto size-32 overflow-hidden rounded-full">
								{#if user.hasImage}
									<!-- Add current timestamp to get around browser caching -->
									<img
										src={`${BASE_URL}users/profile-images/${user.id}?v=${Date.now()}`}
										alt={`${user.firstName} ${user.lastName} profile picture`}
										class="size-full object-cover"
									/>
								{:else}
									<span
										class="flex h-full w-full items-center justify-center
										bg-neutral-200 text-4xl text-neutral-600 dark:bg-neutral-800"
									>
										{user.firstName[0].toUpperCase()}
									</span>
								{/if}
							</div>

							<div class="space-y-2">
								<div>
									<p class="text-center text-lg">{user.firstName} {user.lastName}</p>
									<p class="space-x-2 text-center text-sm">
										{#if user.title}
											<span> {user.title} </span>
											<span> | </span>
										{/if}
										<span>{toTitleCaseDashed(user.role)} </span>
									</p>
								</div>

								<div class="flex justify-between space-y-1 border-t py-4">
									<div class="flex gap-4">
										<div>
											<p class="text-xs text-neutral-600 dark:text-neutral-400">Last Sign-in</p>
											<p class="text-xs">{user.lastSignIn ? formatDate(user.lastSignIn) : 'N/A'}</p>
										</div>
										<div>
											<p class="text-xs text-neutral-600 dark:text-neutral-400">Created On</p>
											<p class="text-xs">{formatDate(user.createdAt)}</p>
										</div>
									</div>
									<div class="flex gap-1">
										{#if user.isActive}
											<div
												class="flex w-fit items-center gap-1 rounded
												bg-green-200 px-2 py-1 text-xs text-green-950
												dark:bg-green-950 dark:text-green-300"
											>
												<Checks />
												<p>Active</p>
											</div>
										{:else}
											<div
												class="flex w-fit items-center gap-1 rounded
												bg-neutral-200 px-2 py-1 text-xs text-neutral-900
												dark:bg-neutral-800 dark:text-neutral-300"
											>
												<Info weight="fill" />
												<p>Inactive</p>
											</div>
										{/if}
										{#if user.isVerified}
											<div
												class="flex w-fit items-center gap-1 rounded
												bg-green-200 px-2 py-1 text-xs text-green-950
												dark:bg-green-950 dark:text-green-300"
											>
												<Checks />
												<p>Verified</p>
											</div>
										{:else}
											<div
												class="flex w-fit items-center gap-1 rounded
												bg-neutral-200 px-2 py-1 text-xs text-neutral-900
												dark:bg-neutral-800 dark:text-neutral-300"
											>
												<Info weight="fill" />
												<p>Pending Verification</p>
											</div>
										{/if}
									</div>
								</div>
							</div>
						</div>

						<div class="border-t py-4">
							{#if user.isActive}
								<button
									onclick={() => openStatusDialog('deactivate')}
									class="cursor-pointer rounded bg-neutral-200 px-4 py-2
									text-xs transition hover:bg-neutral-300 dark:bg-neutral-900
									dark:hover:bg-neutral-800"
								>
									Deactivate
								</button>
							{:else}
								<button
									onclick={() => openStatusDialog('activate')}
									class="cursor-pointer rounded bg-neutral-200 px-4 py-2
									text-xs transition hover:bg-neutral-300 dark:bg-neutral-900
									dark:hover:bg-neutral-800"
								>
									Activate
								</button>
							{/if}
							<button
								onclick={editDialog.open}
								class="cursor-pointer rounded bg-neutral-200 px-4 py-2
								text-xs transition hover:bg-neutral-300 dark:bg-neutral-900
								dark:hover:bg-neutral-800"
							>
								Edit User
							</button>
							<button
								onclick={manageProjectsDialog.open}
								class="cursor-pointer rounded bg-neutral-200 px-4 py-2
								text-xs transition hover:bg-neutral-300 dark:bg-neutral-900
								dark:hover:bg-neutral-800"
							>
								Manage Projects
							</button>

							{#if !user.isVerified}
								<button
									onclick={verifyEmailDialog.open}
									class="cursor-pointer rounded bg-neutral-200 px-4 py-2
									text-xs transition hover:bg-neutral-300 dark:bg-neutral-900
									dark:hover:bg-neutral-800"
								>
									Resend Verification
								</button>
							{/if}
							<button
								onclick={deleteDialog.open}
								class="cursor-pointer rounded bg-red-500 px-4 py-2 text-xs text-red-50
								transition hover:bg-red-600 dark:bg-red-900
								dark:text-red-50 dark:hover:bg-red-800"
							>
								Delete
							</button>
						</div>
					{:else}
						<ErrorMessage variant="warn" text="User Not Found" />
					{/if}
				{:catch error}
					{#if error instanceof ApiError}
						<ErrorMessage variant="warn" text={error.message} />
					{:else}
						<ErrorMessage variant="warn" text="An Error Occurred" />
					{/if}
				{/await}
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

{#if statusAction}
	<UserStatusConfirm
		bind:open={statusDialog.isOpen}
		{userId}
		status={statusAction}
		onSuccess={() => {
			load();
			refresh?.();
		}}
	/>
{/if}

<ResendVerifyEmailConfirm bind:open={verifyEmailDialog.isOpen} {userId} />
<UserDeleteConfirm bind:open={deleteDialog.isOpen} {userId} />
<ManageUserProjects bind:open={manageProjectsDialog.isOpen} {userId} />
<EditUser
	bind:open={editDialog.isOpen}
	{userId}
	onSuccess={() => {
		load();
		refresh?.();
	}}
/>
