<script lang="ts">
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import * as Dialog from '$lib/components/dialogs';
	import * as Table from '$lib/components/ui/table';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import { onDestroy, onMount } from 'svelte';
	import { page } from '$app/state';
	import Spinner from '$lib/components/Spinner.svelte';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';
	import FilterSelect from '$lib/components/FilterSelect.svelte';
	import { formatDate } from '$lib/utils/formatDate';
	import { createDialogState } from '$lib/components/dialogs/createDialogState.svelte';
	import { getUsers, type User } from '$lib/api/users';
	import CheckCircle from 'phosphor-svelte/lib/CheckCircle';
	import DotsThree from 'phosphor-svelte/lib/DotsThree';
	import Trash from 'phosphor-svelte/lib/Trash';
	import { USER_ACTIVE_STATES, USER_ROLES, USER_VERIFIED_STATES } from '$lib/constants/user';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import type { PaginatedResponse } from '$lib/api/page';
	import { ApiError } from '$lib/api/client';

	const MAX_LIMIT = 100;
	const DEFAULT_LIMIT = 25;
	const DEFAULT_PAGE_NUMBER = 1;

	const params = new SvelteURLSearchParams(page.url.searchParams.toString());

	let q = $state(params.get('q') || '');
	let userRole = $state(USER_ROLES.find((r) => r === params.get('role')) ?? '');
	let active = $state(USER_ACTIVE_STATES.find((s) => s === params.get('active')) ?? '');
	let verified = $state(USER_VERIFIED_STATES.find((s) => s === params.get('verified')) ?? '');
	let pageNum = $state(Number(params.get('page')) || DEFAULT_PAGE_NUMBER);

	const limitParam = Number(params.get('limit'));
	let limit = $state(
		Number.isFinite(limitParam)
			? Math.min(Math.max(limitParam, DEFAULT_LIMIT), MAX_LIMIT).toString()
			: DEFAULT_LIMIT.toString()
	);

	let promise: Promise<PaginatedResponse<User>> | null = $state(null);
	let abort: AbortController | null = null;
	function loadUsers() {
		if (abort) {
			abort.abort();
		}

		abort = new AbortController();

		promise = getUsers({ q, role: userRole, limit: Number(limit), page: pageNum }, abort.signal);
	}

	function updateUrlParam() {
		if (q) {
			params.set('q', q);
		} else {
			params.delete('q');
		}

		params.set('page', pageNum.toString());
		params.set('limit', limit.toString());

		if (userRole) {
			params.set('role', userRole);
		} else {
			params.delete('role');
		}

		if (active) {
			params.set('active', active);
		} else {
			params.delete('active');
		}

		if (verified) {
			params.set('verified', verified);
		} else {
			params.delete('verified');
		}

		history.replaceState(null, '', `?${params.toString()}`);
	}

	function handleFilter() {
		pageNum = 1;
		updateUrlParam();
		loadUsers();
	}

	onMount(() => {
		loadUsers();
	});

	onDestroy(() => {
		abort?.abort();
	});

	let newUserDialog = createDialogState();
	let deleteUserDialog = createDialogState();
	let resendVerifyDialog = createDialogState();

	type StatusActionsAllowed = 'activate' | 'deactivate';
	type SelectedUser = User & { action?: StatusActionsAllowed };
	let selectedUser: SelectedUser | null = $state(null);
	let setStatusDialog = createDialogState();

	function openStatusDialog(user: User, action: StatusActionsAllowed) {
		selectedUser = { ...user, action };
		setStatusDialog.open();
	}

	function openDeleteDialog(user: User) {
		selectedUser = { ...user };
		deleteUserDialog.open();
	}

	function openResendVerifyDialog(user: User) {
		selectedUser = { ...user };
		resendVerifyDialog.open();
	}
</script>

<svelte:head>
	<title>Users</title>
</svelte:head>

<div
	class="grid h-full auto-rows-[min-content_1fr_min-content] gap-6 rounded bg-neutral-50 p-4 dark:bg-neutral-950"
>
	<div class="mx-auto flex justify-between lg:container">
		<div class="flex gap-2">
			<div class="w-full max-w-60">
				<SearchBar bind:value={q} onchange={handleFilter} />
			</div>
			<div>
				<FilterSelect
					bind:value={userRole}
					onchange={handleFilter}
					name="Role"
					options={[
						{ value: '', label: 'All' },
						...USER_ROLES.map((s) => ({ value: s, label: toTitleCaseDashed(s) }))
					]}
				/>
			</div>
			<div>
				<FilterSelect
					bind:value={active}
					onchange={handleFilter}
					name="Active"
					options={[
						{ value: '', label: 'All' },
						...USER_ACTIVE_STATES.map((s) => ({ value: s, label: toTitleCaseDashed(s) }))
					]}
				/>
			</div>
			<div>
				<FilterSelect
					bind:value={verified}
					onchange={handleFilter}
					name="Verified"
					options={[
						{ value: '', label: 'All' },
						...USER_VERIFIED_STATES.map((s) => ({ value: s, label: toTitleCaseDashed(s) }))
					]}
				/>
			</div>
			<div>
				<FilterSelect
					bind:value={limit}
					onchange={handleFilter}
					name="Limit"
					options={[
						{ value: '25', label: '25' },
						{ value: '50', label: '50' },
						{ value: '75', label: '75' },
						{ value: '100', label: '100' }
					]}
				/>
			</div>
		</div>

		<button
			onclick={newUserDialog.open}
			class="cursor-pointer rounded bg-neutral-800 px-4 py-2 text-xs text-neutral-50 transition
			hover:bg-neutral-950 dark:bg-neutral-200 dark:text-neutral-950 dark:hover:bg-neutral-50"
		>
			New User
		</button>
	</div>

	{#await promise}
		<Spinner />
	{:then res}
		{#if res && res.items}
			<div class="mx-auto gap-4 overflow-y-auto lg:container">
				{#if res.items.length == 0}
					<ErrorMessage variant="info" text="Change Requests Not Found" />
				{/if}
				<div class="overflow-y-auto">
					{#if res.items.length > 0}
						<Table.Root class="container mx-auto">
							<Table.Header>
								<Table.Row>
									<Table.Head class="font-bold">Name</Table.Head>
									<Table.Head class="font-bold">Title</Table.Head>
									<Table.Head class="font-bold">Role</Table.Head>
									<Table.Head class="font-bold">Active</Table.Head>
									<Table.Head class="font-bold">Verified</Table.Head>
									<Table.Head class="font-bold">Created At</Table.Head>
									<Table.Head class="font-bold">Last Sign-In</Table.Head>
									<Table.Head class="font-bold">Actions</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each res.items as user (user.id)}
									<Table.Row>
										<Table.Cell>{user.firstName} {user.lastName}</Table.Cell>
										<Table.Cell>{user.title ? user.title : 'N/A'}</Table.Cell>
										<Table.Cell>{toTitleCaseDashed(user.role)}</Table.Cell>
										<Table.Cell>{user.isActive ? 'Active' : 'Deactivated'}</Table.Cell>
										<Table.Cell>
											<span class="flex items-center gap-1">
												{#if user.isVerified}
													Verified <CheckCircle weight="fill" size={12} />
												{:else}
													Pending
												{/if}
											</span>
										</Table.Cell>
										<Table.Cell>{formatDate(user.createdAt)}</Table.Cell>
										<Table.Cell>{user.lastLogin ? formatDate(user.lastLogin) : 'N/A'}</Table.Cell>
										<Table.Cell>
											<a
												href={resolve(`/users/${user.id}`)}
												class="inline-block cursor-pointer px-1.5 text-xs
												text-neutral-500 hover:text-neutral-950 dark:text-neutral-400
												dark:hover:text-neutral-50"
												title="View"
											>
												<ArrowRight size={18} />
											</a>

											<DropdownMenu.Root>
												<DropdownMenu.Trigger
													class="cursor-pointer p-1 hover:text-neutral-950 dark:hover:text-neutral-50"
												>
													<DotsThree size={18} />
												</DropdownMenu.Trigger>
												<DropdownMenu.Content class="mr-4 *:text-xs">
													{#if user.isActive}
														<DropdownMenu.Item
															class="text-xs"
															onclick={() => openStatusDialog(user, 'deactivate')}
														>
															Deactivate
														</DropdownMenu.Item>
													{:else}
														<DropdownMenu.Item
															class="text-xs"
															onclick={() => openStatusDialog(user, 'activate')}
														>
															Activate
														</DropdownMenu.Item>
													{/if}
													{#if !user.isVerified}
														<DropdownMenu.Item
															class="text-xs"
															onclick={() => openResendVerifyDialog(user)}
														>
															Resend Verification Email
														</DropdownMenu.Item>
													{/if}
													<DropdownMenu.Item
														class="py-2 text-xs"
														onclick={() => openDeleteDialog(user)}
													>
														<Trash />Delete
													</DropdownMenu.Item>
												</DropdownMenu.Content>
											</DropdownMenu.Root>
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					{/if}
				</div>
			</div>
			{#if res.items.length > 0}
				<div class="container mx-auto flex justify-end">
					<Pagination bind:page={pageNum} count={res.totalCount} perPage={limit} />
				</div>
			{/if}
		{/if}
	{:catch err}
		{#if err instanceof ApiError}
			<ErrorMessage variant="warn" text={err.message} retry={loadUsers} />
		{:else}
			<ErrorMessage variant="warn" text="An Error Occurred" retry={loadUsers} />
		{/if}
	{/await}
</div>

<Dialog.NewUser bind:open={newUserDialog.isOpen} onSuccess={loadUsers} />

{#if selectedUser}
	{#if selectedUser.action}
		<Dialog.UserStatusConfirm
			bind:open={setStatusDialog.isOpen}
			userId={selectedUser.id}
			status={selectedUser.action}
			onSuccess={loadUsers}
		/>
	{/if}

	<Dialog.UserDeleteConfirm
		bind:open={deleteUserDialog.isOpen}
		userId={selectedUser.id}
		onSuccess={loadUsers}
	/>

	<Dialog.ResendVerifyEmailConfirm bind:open={resendVerifyDialog.isOpen} userId={selectedUser.id} />
{/if}
