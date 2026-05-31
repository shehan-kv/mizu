<script lang="ts">
	import { Dialog } from 'bits-ui';
	import X from 'phosphor-svelte/lib/X';
	import SearchUser from '../SearchUser.svelte';
	import UserCard from '../UserCard.svelte';
	import { getProjectMembers, replaceProjectMembers, type ProjectMember } from '$lib/api/projects';
	import { toast } from 'svelte-sonner';
	import Spinner from '../Spinner.svelte';
	import ErrorMessage from '../ErrorMessage.svelte';
	import { ApiError } from '$lib/api/client';

	interface Props {
		open: boolean;
		projectId: string;
		onSuccess?: () => unknown;
	}

	let { open = $bindable(), projectId, onSuccess }: Props = $props();
	let confirmMembers: ProjectMember[] = $state([]);

	let memberLoading = $state(false);
	let memberError: ApiError | null = $state(null);
	let membersAbort: AbortController | null = null;
	async function loadMembers() {
		if (membersAbort) {
			membersAbort.abort();
		}
		membersAbort = new AbortController();

		try {
			memberLoading = true;
			confirmMembers = await getProjectMembers(projectId, membersAbort.signal);
		} catch (error) {
			memberError = error as ApiError;
			confirmMembers = [];
		} finally {
			memberLoading = false;
		}
	}

	function addMember(member: ProjectMember) {
		const exists = confirmMembers.find((m) => m.id == member.id);
		if (!exists) {
			confirmMembers.push(member);
			confirmMembers = [...confirmMembers];
		} else {
			toast.info('Already Added');
		}
	}

	function removeMember(member: ProjectMember) {
		const exists = confirmMembers.find((m) => m.id == member.id);
		if (exists) {
			confirmMembers = confirmMembers.filter((m) => m.id != member.id);
		}
	}

	let confirmAbort: AbortController | null = null;
	async function handleConfirm() {
		if (confirmMembers.length == 0) {
			toast.error('At Least 1 Member Is Required');
			return;
		}

		if (confirmAbort) {
			confirmAbort.abort();
		}
		confirmAbort = new AbortController();

		try {
			await replaceProjectMembers(
				projectId,
				{ memberIds: confirmMembers.map((m) => m.id) },
				confirmAbort.signal
			);
			toast.success('Assigned Members Successfully');
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
			loadMembers();
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
				<p class="font-bold">Manage Members</p>
				<div class="mt-6 space-y-2">
					<div class="h-50 space-y-2 overflow-scroll">
						{#if memberLoading}
							<Spinner />
						{/if}

						{#if memberError}
							<ErrorMessage variant="warn" text={memberError.message} retry={loadMembers} />
						{/if}

						{#if confirmMembers.length > 0}
							{#each confirmMembers as member (member)}
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
					<SearchUser onSelect={(user) => addMember(user)} />
				</div>

				<div class="mt-6 space-x-1 text-right text-xs *:cursor-pointer *:rounded *:px-6 *:py-3">
					<Dialog.Close class="hover:bg-neutral-100 dark:hover:bg-neutral-900">Cancel</Dialog.Close>
					<button
						disabled={confirmMembers.length == 0}
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
