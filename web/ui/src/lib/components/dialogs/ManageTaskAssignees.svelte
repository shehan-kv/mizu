<script lang="ts">
	import { Dialog } from 'bits-ui';
	import X from 'phosphor-svelte/lib/X';
	import { toast } from 'svelte-sonner';
	import Spinner from '../Spinner.svelte';
	import ErrorMessage from '../ErrorMessage.svelte';
	import UserCard from '../UserCard.svelte';
	import SearchProjectMember from '../SearchProjectMember.svelte';
	import type { ProjectMember } from '$lib/api/projects';
	import { ApiError } from '$lib/api/client';
	import { getTaskAssignees, replaceTaskAssignees } from '$lib/api/task';

	interface Props {
		open: boolean;
		projectId: string;
		taskId: string;
		onSuccess: () => unknown;
	}
	let { open = $bindable(), projectId, taskId, onSuccess }: Props = $props();

	let confirmAssignees: ProjectMember[] = $state([]);

	let assigneesLoading = $state(false);
	let assigneesError: ApiError | null = $state(null);
	let assigneesAbort: AbortController | null = null;
	async function loadAssignees() {
		if (assigneesAbort) {
			assigneesAbort.abort();
		}
		assigneesAbort = new AbortController();

		try {
			assigneesLoading = true;
			confirmAssignees = await getTaskAssignees(taskId, assigneesAbort.signal);
		} catch (error) {
			assigneesError = error as ApiError;
			confirmAssignees = [];
		} finally {
			assigneesLoading = false;
		}
	}

	function addAssignee(assignee: ProjectMember) {
		const exists = confirmAssignees.find((a) => a.id == assignee.id);
		if (!exists) {
			confirmAssignees.push(assignee);
			confirmAssignees = [...confirmAssignees];
		} else {
			toast.info('Already Added');
		}
	}

	function removeAssignee(assignee: ProjectMember) {
		const exists = confirmAssignees.find((a) => a.id == assignee.id);
		if (exists) {
			confirmAssignees = confirmAssignees.filter((a) => a.id != assignee.id);
		}
	}

	let confirmAbort: AbortController | null = null;
	async function handleConfirm() {
		if (confirmAssignees.length == 0) {
			toast.error('At Least 1 Assignee Is Required');
			return;
		}

		if (confirmAbort) {
			confirmAbort.abort();
		}
		confirmAbort = new AbortController();

		try {
			await replaceTaskAssignees(
				taskId,
				{ assigneeIds: confirmAssignees.map((a) => a.id) },
				confirmAbort.signal
			);
			toast.success('Assigned Successfully');
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

	$effect(() => {
		if (open) {
			loadAssignees();
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
				<p class="font-bold">Manage Assignees</p>
				<div class="mt-6 space-y-2">
					<div class="h-50 space-y-2 overflow-scroll">
						{#if assigneesLoading}
							<Spinner />
						{/if}

						{#if assigneesError}
							<ErrorMessage variant="warn" text={assigneesError.message} retry={loadAssignees} />
						{/if}

						{#if confirmAssignees.length > 0}
							{#each confirmAssignees as assignee (assignee)}
								<div class="grid grid-cols-[1fr_min-content] items-center">
									<UserCard
										image={assignee.image}
										role={assignee.role}
										title={assignee.title}
										name={`${assignee.firstName} ${assignee.lastName}`}
									/>

									<button
										onclick={() => removeAssignee(assignee)}
										class="cursor-pointer rounded p-2 transition hover:bg-neutral-900"
									>
										<X />
									</button>
								</div>
							{/each}
						{/if}
					</div>
					<SearchProjectMember {projectId} onSelect={(user) => addAssignee(user)} />
				</div>

				<div class="mt-6 space-x-1 text-right text-xs *:cursor-pointer *:rounded *:px-6 *:py-3">
					<Dialog.Close class="hover:bg-neutral-100 dark:hover:bg-neutral-900">Cancel</Dialog.Close>
					<button
						disabled={confirmAssignees.length == 0}
						onclick={handleConfirm}
						class="bg-neutral-800 text-neutral-50 transition hover:bg-neutral-950
								disabled:cursor-not-allowed dark:bg-neutral-200 dark:text-neutral-950
								dark:hover:bg-neutral-50"
					>
						Confirm
					</button>
				</div>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
