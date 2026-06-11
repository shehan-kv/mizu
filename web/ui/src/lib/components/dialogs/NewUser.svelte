<script lang="ts">
	import X from 'phosphor-svelte/lib/X';
	import { Dialog, Select } from 'bits-ui';
	import Check from 'phosphor-svelte/lib/Check';
	import CaretUpDown from 'phosphor-svelte/lib/CaretUpDown';
	import CaretDoubleUp from 'phosphor-svelte/lib/CaretDoubleUp';
	import CaretDoubleDown from 'phosphor-svelte/lib/CaretDoubleDown';
	import { toast } from 'svelte-sonner';
	import InputLabel from '../InputLabel.svelte';
	import ErrorMessage from '../ErrorMessage.svelte';
	import Info from 'phosphor-svelte/lib/Info';
	import { createUser, type UserRole } from '$lib/api/users';
	import { type ProjectStat } from '$lib/api/projects';
	import SearchProject from '../SearchProject.svelte';
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';
	import { ApiError } from '$lib/api/client';

	interface Props {
		open: boolean;
		onSuccess?: () => unknown;
	}

	let { open = $bindable(), onSuccess }: Props = $props();

	const roles: { value: UserRole; label: string }[] = [
		{ value: 'administrator', label: 'Administrator' },
		{ value: 'staff', label: 'Staff' },
		{ value: 'client', label: 'Client' }
	];

	type UserStatus = 'active' | 'disabled';

	const statuses: { value: UserStatus; label: string }[] = [
		{ value: 'active', label: 'Active' },
		{ value: 'disabled', label: 'Disabled' }
	];

	let req = $state<{
		firstName: string;
		lastName: string;
		title: string;
		email: string;
		role: UserRole;
		status: UserStatus;
		projects: ProjectStat[];
		projectSearchTerm: string;
	}>({
		firstName: '',
		lastName: '',
		title: '',
		email: '',
		role: roles[0].value,
		status: statuses[0].value,
		projects: [],
		projectSearchTerm: ''
	});

	function resetState() {
		req.firstName = '';
		req.lastName = '';
		req.title = '';
		req.email = '';
		req.role = roles[0].value;
		req.status = statuses[0].value;
		req.projects = [];
		req.projectSearchTerm = '';
	}

	function validateRequest() {
		return req.firstName.trim() !== '' && req.lastName.trim() !== '' && req.email.trim() !== '';
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

			await createUser(
				{
					firstName: req.firstName,
					lastName: req.lastName,
					email: req.email,
					role: req.role,
					isActive: req.status == 'active' ? true : false,
					projectIds: req.projects.map((p) => p.id)
				},
				createAbort.signal
			);

			toast.success('Successfully Created');
			resetState();
			onSuccess?.();
			open = false;
		} catch (error) {
			if (error instanceof ApiError) {
				toast.error(error.message);
			} else {
				toast.error('An Error Occurred');
			}
		}
	}

	function addProject(project: ProjectStat) {
		const exists = req.projects.find((p) => p.id == project.id);
		if (!exists) {
			req.projects.push(project);
			req.projects = [...req.projects];
		}
	}

	function removeProject(project: ProjectStat) {
		const exists = req.projects.find((p) => p.id == project.id);
		if (exists) {
			req.projects = req.projects.filter((p) => p.id != project.id);
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
				<p class="font-bold">Create New User</p>
				<div class="mt-6 grid grid-cols-2 gap-2 space-y-2">
					<div class="col-span-2 grid grid-cols-3 gap-2">
						<div class="space-y-1 text-sm *:block">
							<InputLabel htmlFor="firstName" text="First Name" required />
							<input
								type="text"
								id="firstName"
								bind:value={req.firstName}
								class="w-full rounded border border-neutral-200 bg-neutral-100 p-2
								outline-hidden dark:border-neutral-800 dark:bg-neutral-900"
							/>
						</div>
						<div class="space-y-1 text-sm *:block">
							<InputLabel htmlFor="lastName" text="Last Name" required />
							<input
								type="text"
								id="lastName"
								bind:value={req.lastName}
								class="w-full rounded border border-neutral-200 bg-neutral-100 p-2
                        		outline-hidden dark:border-neutral-800 dark:bg-neutral-900"
							/>
						</div>
						<div class="space-y-1 text-sm *:block">
							<InputLabel htmlFor="title" text="Title" />
							<input
								type="text"
								id="title"
								bind:value={req.title}
								class="w-full rounded border border-neutral-200 bg-neutral-100 p-2
                        		outline-hidden dark:border-neutral-800 dark:bg-neutral-900"
							/>
						</div>
					</div>
					<div class="grow space-y-1 text-sm *:block">
						<InputLabel htmlFor="email" text="Email" required />
						<input
							type="email"
							id="email"
							bind:value={req.email}
							class="w-full rounded border border-neutral-200 bg-neutral-100 p-2
                            outline-hidden dark:border-neutral-800 dark:bg-neutral-900"
						/>
					</div>
					<div class="grid grid-cols-2 gap-2">
						<div class="space-y-1 text-sm *:block">
							<InputLabel htmlFor="role" text="Role" />
							<Select.Root type="single" bind:value={req.role} items={roles} allowDeselect={false}>
								<Select.Trigger
									class="w-32 gap-2 rounded border border-neutral-200 
                                    bg-neutral-100 px-2 py-2.5 dark:border-neutral-800 dark:bg-neutral-900"
									aria-label="Select a status"
								>
									<p class="flex items-center text-xs">
										{roles.find((role) => role.value === req.role)?.label}
										<CaretUpDown class="text-muted-foreground ml-auto size-3" />
									</p>
								</Select.Trigger>

								<Select.Portal>
									<Select.Content
										class="focus-override border-muted bg-background shadow-popover data-[state=open]:animate-in 
                                        data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 
                                        data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 
                                        data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 
                                        data-[side=top]:slide-in-from-bottom-2 z-50 max-h-(--bits-select-content-available-height) 
                                        w-fit min-w-(--bits-select-anchor-width) rounded-xl 
                                        border px-1 py-3 outline-hidden select-none data-[side=bottom]:translate-y-1 
                                        data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 data-[side=top]:-translate-y-1"
										sideOffset={10}
									>
										<Select.ScrollUpButton class="flex w-full items-center justify-center">
											<CaretDoubleUp class="size-3" />
										</Select.ScrollUpButton>
										<Select.Viewport class="p-1">
											{#each roles as role, i (i + role.value)}
												<Select.Item
													class="data-highlighted:bg-muted flex 
                                                        h-fit w-full items-center 
                                                        gap-1 rounded px-4 py-2 text-xs 
                                                        capitalize outline-hidden select-none data-disabled:opacity-50"
													value={role.value}
													label={role.label}
												>
													{#snippet children({ selected })}
														{role.label}
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
                                        data-[side=top]:slide-in-from-bottom-2 z-50 max-h-(--bits-select-content-available-height) 
                                        w-fit min-w-(--bits-select-anchor-width) rounded-xl 
                                        border px-1 py-3 outline-hidden select-none data-[side=bottom]:translate-y-1 
                                        data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 data-[side=top]:-translate-y-1"
										sideOffset={10}
									>
										<Select.ScrollUpButton class="flex w-full items-center justify-center">
											<CaretDoubleUp class="size-3" />
										</Select.ScrollUpButton>
										<Select.Viewport class="p-1">
											{#each statuses as status, i (i + status.value)}
												<Select.Item
													class="data-highlighted:bg-muted flex 
                                                        h-fit w-full items-center 
                                                        gap-1 rounded px-4 py-2 text-xs 
                                                        capitalize outline-hidden select-none data-disabled:opacity-50"
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
				</div>
				<div class="mt-4">
					<p class="text-sm">Assign Projects</p>
					<div>
						<div class="h-30 space-y-2 overflow-scroll py-4">
							{#if req.projects.length == 0}
								<ErrorMessage variant="info" text="No Projects Assigned" />
							{:else}
								{#each req.projects as project (project)}
									<div class="grid grid-cols-[1fr_min-content] items-center">
										<div>
											<p class="text-sm">{project.name}</p>
											<p class="text-xs text-neutral-400">{toTitleCaseDashed(project.status)}</p>
										</div>

										<button
											onclick={() => removeProject(project)}
											class="cursor-pointer rounded p-2 transition hover:bg-neutral-900"
										>
											<X />
										</button>
									</div>
								{/each}
							{/if}
						</div>

						<div class="space-y-2 border-t py-2">
							<SearchProject onSelect={addProject} />

							<div class="space-y-2 text-xs text-neutral-500">
								<div class="flex items-center gap-1">
									<Info size={16} />
									<p>Assigning Projects Is Optional</p>
								</div>
								<p>
									Upon Account Creation, The User Will Receive An Email To Confirm The Account And
									Set A Password.
								</p>
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
