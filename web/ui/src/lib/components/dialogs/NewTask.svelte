<script lang="ts">
	import {
		APIBadRequestError,
		APIError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError,
		NetworkError
	} from '$lib/api/errors';
	import { createTask, deleteProject, type ProjectMember } from '$lib/api/projects';
	import { Dialog, Select } from 'bits-ui';
	import CaretDoubleDown from 'phosphor-svelte/lib/CaretDoubleDown';
	import CaretDoubleUp from 'phosphor-svelte/lib/CaretDoubleUp';
	import CaretUpDown from 'phosphor-svelte/lib/CaretUpDown';
	import Check from 'phosphor-svelte/lib/Check';
	import X from 'phosphor-svelte/lib/X';
	import { toast } from 'svelte-sonner';
	import InputLabel from '../InputLabel.svelte';
	import SearchUser from '../SearchUser.svelte';
	import UserCard from '../UserCard.svelte';
	import ErrorMessage from '../ErrorMessage.svelte';

	interface Props {
		open: boolean;
		projectId: number;
		onSuccess?: () => any;
	}

	let { open = $bindable(), projectId, onSuccess }: Props = $props();

	let req = $state<{
		name: string;
		description: string;
		priority: 'high' | 'medium' | 'low';
		status: 'backlog';
		estimatedTimeUnit: 'minutes' | 'hours' | 'days' | 'weeks' | 'years';
		estimatedTime: number;
		assignees: ProjectMember[];
	}>({
		name: '',
		description: '',
		priority: 'high',
		status: 'backlog',
		estimatedTimeUnit: 'minutes',
		estimatedTime: 0,
		assignees: []
	});

	function resetState() {
		req = {
			name: '',
			description: '',
			priority: 'high',
			status: 'backlog',
			estimatedTimeUnit: 'minutes',
			estimatedTime: 0,
			assignees: []
		};
	}

	function addAssignee(assignee: ProjectMember) {
		const exists = req.assignees.find((a) => a.id == assignee.id);
		if (!exists) {
			req.assignees.push(assignee);
			req.assignees = [...req.assignees];
		} else {
			toast.info('Already Added');
		}
	}

	function removeAssignee(assignee: ProjectMember) {
		const exists = req.assignees.find((a) => a.id == assignee.id);
		if (exists) {
			req.assignees = req.assignees.filter((a) => a.id != assignee.id);
		}
	}

	const priorities = [
		{ value: 'high', label: 'High' },
		{ value: 'medium', label: 'Medium' },
		{ value: 'low', label: 'Low' }
	];

	const timeUnits = [
		{ value: 'minutes', label: 'Minutes' },
		{ value: 'hours', label: 'Hours' },
		{ value: 'days', label: 'Days' },
		{ value: 'weeks', label: 'Weeks' },
		{ value: 'months', label: 'Months' },
		{ value: 'years', label: 'Years' }
	];

	const statuses = [
		{ value: 'backlog', label: 'Backlog' },
		{ value: 'in-progress', label: 'In-Progress' },
		{ value: 'completed', label: 'Completed' }
	];

	function timeUnitsToMinutes(time: number) {
		const conversionRates = {
			minutes: 1,
			hours: 60,
			days: 1440,
			weeks: 10080,
			months: 43800,
			years: 525600
		};

		if (!conversionRates[req.estimatedTimeUnit]) {
			throw new Error('Unsupported time unit');
		}

		return time * conversionRates[req.estimatedTimeUnit];
	}

	let createAbort: AbortController | null = null;
	async function handleDelete() {
		if (createAbort) {
			createAbort.abort();
		}

		createAbort = new AbortController();

		try {
			await createTask(
				projectId,
				{
					name: req.name,
					description: req.description,
					priority: req.priority,
					status: req.status,
					assignees: req.assignees.map((a) => a.id),
					estimatedTimeMinutes: timeUnitsToMinutes(req.estimatedTime)
				},
				createAbort.signal
			);

			toast.success('Successfully Created');
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
			} else {
				toast.error('Unexpected Error, Try Again');
			}
		}
	}
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
				<p class="font-bold">Create New Task</p>

				<div class="mt-4 text-sm">
					<div class="grid grid-cols-3 gap-3">
						<div class="space-y-1">
							<p>Priority</p>
							<Select.Root
								type="single"
								bind:value={req.priority}
								items={priorities}
								allowDeselect={false}
							>
								<Select.Trigger
									class="inline-flex w-full items-center gap-2 rounded bg-neutral-100 px-4 py-2.5 text-xs dark:bg-neutral-900"
									aria-label="Select priority"
								>
									{priorities.find((p) => p.value == req.priority)?.label}
									<CaretUpDown class="text-muted-foreground ml-auto size-3" />
								</Select.Trigger>
								<Select.Portal>
									<Select.Content
										class="focus-override border-muted bg-background shadow-popover data-[state=open]:animate-in 
									   data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 
									data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 
									data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 
									data-[side=top]:slide-in-from-bottom-2 outline-hidden max-h-100 z-50
									w-fit min-w-[var(--bits-select-anchor-width)] select-none rounded-xl border px-1 py-3 
									data-[side=bottom]:translate-y-1 data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 
									data-[side=top]:-translate-y-1"
										sideOffset={10}
									>
										<Select.ScrollUpButton class="flex w-full items-center justify-center">
											<CaretDoubleUp class="size-3" />
										</Select.ScrollUpButton>
										<Select.Viewport class="p-1">
											{#each priorities as option, i (i + option.value)}
												<Select.Item
													class="data-highlighted:bg-muted outline-hidden data-disabled:opacity-50 flex h-fit 
												w-full select-none items-center gap-1 rounded px-4 py-2 text-xs capitalize"
													value={option.value}
													label={option.label}
												>
													{#snippet children({ selected })}
														{option.label}
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

						<div class="space-y-1">
							<p>Status</p>
							<Select.Root
								type="single"
								bind:value={req.status}
								items={statuses}
								allowDeselect={false}
							>
								<Select.Trigger
									class="inline-flex w-full items-center gap-2 rounded bg-neutral-100 px-4 py-2.5 text-xs dark:bg-neutral-900"
									aria-label="Select priority"
								>
									{statuses.find((s) => s.value == req.status)?.label}
									<CaretUpDown class="text-muted-foreground ml-auto size-3" />
								</Select.Trigger>
								<Select.Portal>
									<Select.Content
										class="focus-override border-muted bg-background shadow-popover data-[state=open]:animate-in 
									   data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 
									data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 
									data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 
									data-[side=top]:slide-in-from-bottom-2 outline-hidden max-h-100 z-50
									w-fit min-w-[var(--bits-select-anchor-width)] select-none rounded-xl border px-1 py-3 
									data-[side=bottom]:translate-y-1 data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 
									data-[side=top]:-translate-y-1"
										sideOffset={10}
									>
										<Select.ScrollUpButton class="flex w-full items-center justify-center">
											<CaretDoubleUp class="size-3" />
										</Select.ScrollUpButton>
										<Select.Viewport class="p-1">
											{#each statuses as option, i (i + option.value)}
												<Select.Item
													class="data-highlighted:bg-muted outline-hidden data-disabled:opacity-50 flex h-fit 
												w-full select-none items-center gap-1 rounded px-4 py-2 text-xs capitalize"
													value={option.value}
													label={option.label}
												>
													{#snippet children({ selected })}
														{option.label}
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

						<div class="space-y-1">
							<p>Estimated Time</p>
							<div class="grid grid-cols-[1fr_min-content]">
								<input
									required
									bind:value={req.estimatedTime}
									min="0"
									type="number"
									name="estimateTime"
									id="estimatedTime"
									placeholder="Enter Time"
									class="w-full rounded-l bg-neutral-100 pl-4 [appearance:textfield] placeholder:text-xs
									placeholder:italic dark:bg-neutral-900 [&::-webkit-inner-spin-button]:appearance-none
									[&::-webkit-outer-spin-button]:appearance-none"
								/>
								<Select.Root
									type="single"
									bind:value={req.estimatedTimeUnit}
									items={timeUnits}
									allowDeselect={false}
								>
									<Select.Trigger
										class="inline-flex items-center gap-2 rounded-r bg-neutral-100 px-4 py-2.5 text-xs dark:bg-neutral-900"
										aria-label="Select priority"
									>
										{timeUnits.find((t) => t.value == req.estimatedTimeUnit)?.label}
										<CaretUpDown class="text-muted-foreground ml-auto size-3" />
									</Select.Trigger>
									<Select.Portal>
										<Select.Content
											class="focus-override border-muted bg-background shadow-popover data-[state=open]:animate-in 
									   data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 
									data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 
									data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 
									data-[side=top]:slide-in-from-bottom-2 outline-hidden max-h-100 z-50
									w-fit min-w-[var(--bits-select-anchor-width)] select-none rounded-xl border px-1 py-3 
									data-[side=bottom]:translate-y-1 data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 
									data-[side=top]:-translate-y-1"
											sideOffset={10}
										>
											<Select.ScrollUpButton class="flex w-full items-center justify-center">
												<CaretDoubleUp class="size-3" />
											</Select.ScrollUpButton>
											<Select.Viewport class="p-1">
												{#each timeUnits as option, i (i + option.value)}
													<Select.Item
														class="data-highlighted:bg-muted outline-hidden data-disabled:opacity-50 flex h-fit 
										w-full select-none items-center gap-1 rounded px-4 py-2 text-xs capitalize"
														value={option.value}
														label={option.label}
													>
														{#snippet children({ selected })}
															{option.label}
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
					</div>

					<div class="mt-4">
						<InputLabel htmlFor="name" text="Name" required />
						<input
							type="text"
							id="name"
							bind:value={req.name}
							class="outline-hidden mt-1 block w-full rounded border border-neutral-200 bg-neutral-100
							p-2 dark:border-neutral-800 dark:bg-neutral-900"
						/>
					</div>

					<div class="mt-4 space-y-1">
						<InputLabel htmlFor="description" text="Description" />
						<textarea
							bind:value={req.description}
							name="description"
							id="description"
							class="h-20 w-full resize-none rounded bg-neutral-100 p-4 dark:bg-neutral-900"
						></textarea>
					</div>

					<div class="mt-4">
						<p>Assignees</p>
						<div class="h-30 space-y-2 overflow-scroll py-2">
							{#if req.assignees.length > 0}
								{#each req.assignees as user}
									<div class="flex items-center justify-between gap-2">
										<UserCard
											image={user.image}
											name={`${user.firstName} ${user.lastName}`}
											role={user.role}
											title={user.title}
										/>

										<button
											class="cursor-pointer rounded p-2 transition hover:bg-neutral-100
											dark:hover:bg-neutral-900"
											onclick={() => removeAssignee(user)}
										>
											<X size={18} />
										</button>
									</div>
								{/each}
							{:else}
								<ErrorMessage text="No Assignees Yet" variant="info" />
							{/if}
						</div>
					</div>
					<div>
						<SearchUser onSelect={addAssignee} />
					</div>
				</div>

				<div class="mt-6 space-x-1 text-right text-xs *:cursor-pointer *:rounded *:px-6 *:py-3">
					<Dialog.Close class="hover:bg-neutral-100 dark:hover:bg-neutral-900">Cancel</Dialog.Close>
					<button
						onclick={handleDelete}
						class="bg-neutral-800 text-neutral-50 hover:bg-neutral-950 dark:bg-neutral-200
						dark:text-neutral-950 dark:hover:bg-neutral-50"
					>
						Create
					</button>
				</div>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
