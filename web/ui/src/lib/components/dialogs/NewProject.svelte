<script lang="ts">
	import X from 'phosphor-svelte/lib/X';
	import { Dialog, Select } from 'bits-ui';
	import Check from 'phosphor-svelte/lib/Check';
	import CaretUpDown from 'phosphor-svelte/lib/CaretUpDown';
	import CaretDoubleUp from 'phosphor-svelte/lib/CaretDoubleUp';
	import CaretDoubleDown from 'phosphor-svelte/lib/CaretDoubleDown';
	import { toast } from 'svelte-sonner';
	import InputLabel from '../InputLabel.svelte';
	import {
		APIBadRequestError,
		APIError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError,
		NetworkError
	} from '$lib/api/errors';
	import ErrorMessage from '../ErrorMessage.svelte';
	import Info from 'phosphor-svelte/lib/Info';
	import MagnifyingGlass from 'phosphor-svelte/lib/MagnifyingGlass';
	import Spinner from '../Spinner.svelte';
	import { debounce } from '$lib/utils/debounce';
	import { getUsers, type User } from '$lib/api/users';
	import UserCard from '../UserCard.svelte';
	import { createProject, type ProjectStatus } from '$lib/api/projects';

	interface Props {
		open: boolean;
		onSuccess?: () => any;
	}

	let { open = $bindable(), onSuccess }: Props = $props();

	const statuses: { value: ProjectStatus; label: string }[] = [
		{ value: 'started', label: 'Started' },
		{ value: 'paused', label: 'Paused' },
		{ value: 'cancelled', label: 'Cancelled' },
		{ value: 'completed', label: 'Completed' }
	];

	let req = $state<{
		name: string;
		status: ProjectStatus;
		members: User[];
		memberSearchTerm: string;
	}>({
		name: '',
		status: statuses[0].value,
		members: [],
		memberSearchTerm: ''
	});

	function resetState() {
		req.name = '';
		req.status = statuses[0].value;
		req.members = [];
		req.memberSearchTerm = '';
	}

	function validateRequest() {
		return req.name.trim() !== '';
	}

	let createAbort: AbortController | null = null;
	async function handleCreate() {
		try {
			if (createAbort) {
				createAbort.abort();
			}

			createAbort = new AbortController();

			if (!validateRequest()) {
				toast.error('All Fields Are Required');
				return;
			}

			await createProject(
				{
					name: req.name,
					status: req.status,
					members: req.members.map((m) => m.id)
				},
				createAbort.signal
			);

			toast.success('Successfully Created');
			resetState();
			onSuccess && onSuccess();
			open = false;
		} catch (error) {
			if (error instanceof APIBadRequestError) {
				toast.error('Invalid Request');
			} else if (error instanceof APIForbiddenError) {
				toast.error('Not Authorized');
			} else if (error instanceof APINotFoundError) {
				toast.error('Not Found');
			} else if (error instanceof APIServerError) {
				toast.error('Server Error');
			} else if (error instanceof APIError) {
				toast.error('Unexpected Error, Try Again');
			} else if (error instanceof NetworkError) {
				toast.error('Request Failed, Try Again');
			}
		}
	}

	let membersPromise: Promise<PaginatedResponse<User>> | null = $state(null);
	let membersAbort: AbortController | null = null;
	function loadMembers() {
		if (!req.memberSearchTerm) {
			if (membersAbort) {
				membersAbort.abort();
				membersAbort = null;
			}

			membersPromise = null;
			return;
		}

		if (membersAbort) {
			membersAbort.abort();
		}

		membersAbort = new AbortController();

		membersPromise = getUsers(
			{ q: req.memberSearchTerm.trim(), page: 1, limit: 50 },
			membersAbort.signal
		);
	}

	function addMember(member: User) {
		const exists = req.members.find((m) => m.id == member.id);
		if (!exists) {
			req.members.push(member);
			req.members = [...req.members];
		}
	}

	function removeMember(member: User) {
		const exists = req.members.find((m) => m.id == member.id);
		if (exists) {
			req.members = req.members.filter((m) => m.id != member.id);
		}
	}

	const memberSearchDebounced = debounce(() => {
		loadMembers();
	}, 300);
</script>

<Dialog.Root bind:open onOpenChange={resetState}>
	<Dialog.Portal>
		<Dialog.Overlay
			class="data-[state=open]:animate-in data-[state=closed]:animate-out 
			data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed 
			inset-0 z-50 bg-neutral-100/80 dark:bg-black/80"
		/>
		<Dialog.Content
			class="bg-background data-[state=open]:animate-in data-[state=closed]:animate-out 
			data-[state=closed]:slide-out-to-bottom-8 data-[state=closed]:fade-out
			data-[state=open]:slide-in-from-bottom-8 data-[state=open]:fade-in 
			outline-hidden duration-250 fixed left-1/2 top-1/2 z-50 grid w-full max-w-xl -translate-x-1/2 -translate-y-1/2 auto-rows-[min-content_1fr] gap-4 
			rounded"
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
				<p class="font-bold">Create New Project</p>
				<div class="mt-6 flex gap-2 space-y-4">
					<div class="grow space-y-1 text-sm *:block">
						<InputLabel htmlFor="name" text="Name" required />
						<input
							type="text"
							id="name"
							bind:value={req.name}
							class="outline-hidden w-full rounded border border-neutral-200 bg-neutral-100
							p-2 dark:border-neutral-800 dark:bg-neutral-900"
						/>
					</div>
					<div class="space-y-1 text-sm *:block">
						<InputLabel htmlFor="status" text="Status" />
						<Select.Root
							type="single"
							bind:value={req.status}
							items={statuses}
							allowDeselect={false}
						>
							<Select.Trigger
								class="w-32 gap-2 rounded border border-neutral-200 
								bg-neutral-100 px-2 py-2.5 dark:border-neutral-800 dark:bg-neutral-900"
								aria-label="Select a status"
							>
								<p class="flex items-center text-xs">
									{statuses.find((status) => status.value === req.status)?.label}
									<CaretUpDown class="text-muted-foreground ml-auto size-3" />
								</p>
							</Select.Trigger>

							<Select.Portal>
								<Select.Content
									class="focus-override border-muted bg-background shadow-popover data-[state=open]:animate-in 
            						data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 
            						data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 
            						data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 
            						data-[side=top]:slide-in-from-bottom-2 outline-hidden z-50 
									max-h-[var(--bits-select-content-available-height)] w-fit min-w-[var(--bits-select-anchor-width)] 
									select-none rounded-xl border px-1 py-3 data-[side=bottom]:translate-y-1 
									data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 data-[side=top]:-translate-y-1"
									sideOffset={10}
								>
									<Select.ScrollUpButton class="flex w-full items-center justify-center">
										<CaretDoubleUp class="size-3" />
									</Select.ScrollUpButton>
									<Select.Viewport class="p-1">
										{#each statuses as status, i (i + status.value)}
											<Select.Item
												class="data-highlighted:bg-muted outline-hidden 
													data-disabled:opacity-50 flex h-fit 
                        							w-full select-none items-center gap-1 rounded 
													px-4 py-2 text-xs capitalize"
												value={status.value}
												label={status.label}
											>
												{#snippet children({ selected })}
													{status.label}
													{#if selected}
														<div class="ml-auto">
															<Check aria-label="check" />
														</div>
													{/if}
												{/snippet}
											</Select.Item>
										{/each}
									</Select.Viewport>
									<Select.ScrollDownButton class="flex w-full items-center justify-center">
										<CaretDoubleDown class="size-3" />
									</Select.ScrollDownButton>
								</Select.Content>
							</Select.Portal>
						</Select.Root>
					</div>
				</div>
				<div class="mt-2">
					<p class="text-sm">Assign Other Members</p>
					<div>
						<div class="h-30 space-y-2 overflow-scroll py-4">
							{#if req.members.length == 0}
								<ErrorMessage variant="info" text="No Other Members Assigned" />
							{:else}
								{#each req.members as member (member)}
									<div class="grid grid-cols-[1fr_min-content] items-center">
										<UserCard
											image={member.image}
											role={member.role}
											title={member.title}
											name={`${member.firstName} ${member.lastName}`}
										/>

										<button
											onclick={() => removeMember(member)}
											class="cursor-pointer rounded p-2 transition hover:bg-neutral-900"
										>
											<X />
										</button>
									</div>
								{/each}
							{/if}
						</div>

						<div class="space-y-2 border-t py-2">
							<div
								class="relative flex items-center gap-1 rounded border
									border-neutral-200 bg-neutral-100
									dark:border-neutral-800 dark:bg-neutral-900 focus-within:[&>div.absolute]:block"
							>
								<input
									bind:value={req.memberSearchTerm}
									oninput={memberSearchDebounced}
									type="text"
									class="outline-hidden peer grow p-2 text-sm placeholder:text-xs placeholder:italic"
									placeholder="Search For Members..."
								/>
								<MagnifyingGlass size={16} class="mx-2" />

								<div
									class="max-h-50 absolute left-0 top-10 hidden min-h-10 w-full overflow-scroll
									rounded bg-neutral-900 px-2 py-3 ring-0 transition"
								>
									{#if !req.memberSearchTerm && !membersPromise}
										<p class="text-xs text-neutral-300">Start Typing To Search</p>
									{/if}

									{#await membersPromise}
										<Spinner size={16} />
									{:then res}
										{#if res && res.data.length > 0}
											<div>
												{#each res.data as member}
													<div
														class="cursor-pointer rounded p-2 hover:bg-neutral-950"
														onmousedown={() => {
															console.log('clicked');
															addMember(member);
														}}
														role="button"
														tabindex="0"
														onkeydown={(e) => {
															if (e.key === 'Enter' || e.key === ' ') {
																e.preventDefault();
																addMember(member);
															}
														}}
													>
														<UserCard
															image={member.image}
															role={member.role}
															title={member.title}
															name={`${member.firstName} ${member.lastName}`}
														/>
													</div>
												{/each}
											</div>
										{:else if res && req.memberSearchTerm}
											<ErrorMessage variant="info" text="Members Not Found" />
										{/if}
									{/await}
								</div>
							</div>
							<div class="flex items-center gap-1 text-xs text-neutral-500">
								<Info size={16} />
								<p>You Are Automatically Added As A Member</p>
							</div>
						</div>
					</div>
				</div>
				<div class="mt-6 space-x-1 text-right text-xs *:cursor-pointer *:rounded *:px-6 *:py-3">
					<Dialog.Close class="hover:bg-neutral-100 dark:hover:bg-neutral-900">Cancel</Dialog.Close>
					<button
						onclick={handleCreate}
						class="bg-neutral-800 text-neutral-50 transition hover:bg-neutral-950
                    dark:bg-neutral-200 dark:text-neutral-950 dark:hover:bg-neutral-50"
					>
						Create
					</button>
				</div>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
