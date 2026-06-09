<script lang="ts">
	import { Dialog } from 'bits-ui';
	import X from 'phosphor-svelte/lib/X';
	import InputLabel from '../InputLabel.svelte';
	import ErrorMessage from '../ErrorMessage.svelte';
	import SearchProject from '../SearchProject.svelte';
	import type { ProjectStat } from '$lib/api/projects';
	import { createChannel } from '$lib/api/messages';
	import SearchUser from '../SearchUser.svelte';
	import type { User } from '$lib/api/users';
	import UserCard from '../UserCard.svelte';
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';
	import { toast } from 'svelte-sonner';
	import { ApiError } from '$lib/api/client';
	import Info from 'phosphor-svelte/lib/Info';

	interface Props {
		open: boolean;
	}

	let { open = $bindable() }: Props = $props();

	let req = $state<{
		name: string;
		project: ProjectStat | null;
		members: User[];
	}>({
		name: '',
		project: null,
		members: []
	});

	function resetState() {
		req = {
			name: '',
			project: null,
			members: []
		};
	}

	function addProject(project: ProjectStat) {
		req.project = project;
	}

	function addMember(user: User) {
		req.members = [...req.members, user];
	}

	function removeMember(user: User) {
		req.members = req.members.filter((m) => m.id != user.id);
	}

	function validateRequest() {
		return req.name.trim() != '' && req.members.length >= 1;
	}

	let abort: AbortController | null = null;
	async function handleCreate() {
		try {
			if (abort) {
				abort.abort();
			}

			abort = new AbortController();

			if (!validateRequest()) {
				toast.error('Missing Required Information');
				return;
			}

			await createChannel(
				{
					name: req.name,
					projectId: req.project?.id,
					memberIds: req.members.map((m) => m.id)
				},
				abort.signal
			);

			toast.success('Successfully Created');
			resetState();
			open = false;
		} catch (error) {
			if (error instanceof ApiError) {
				toast.error(error.message);
			} else {
				toast.error('An Error Occurred');
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
				<p class="font-bold">Create New Channel</p>
				<div class="mt-6 space-y-4">
					<div>
						<div class="space-y-1 text-sm *:block">
							<InputLabel htmlFor="name" text="Channel Name" required />
							<input
								type="text"
								id="name"
								bind:value={req.name}
								class="w-full rounded border border-neutral-200 bg-neutral-100 p-2
								outline-hidden dark:border-neutral-800 dark:bg-neutral-900"
							/>
						</div>
					</div>
					<div class="space-y-2 border-t py-4">
						<p class="text-sm">Assign A Project (Optional)</p>
						<div class="space-y-4 text-sm">
							{#if !req.project}
								<p class="text-neutral-400">No Project Assigned</p>
							{:else}
								<div class="grid grid-cols-[1fr_min-content] items-center gap-2">
									<div class="overflow-hidden">
										<p class="truncate text-ellipsis text-neutral-400">{req.project.name}</p>
										<p class="text-xs text-neutral-400">
											{toTitleCaseDashed(req.project?.status ?? 'status here')}
										</p>
									</div>

									<button
										title="Remove Project"
										class="cursor-pointer rounded p-2 transition hover:bg-neutral-900"
									>
										<X />
									</button>
								</div>
							{/if}

							<div>
								<SearchProject onSelect={addProject} />
							</div>
						</div>
					</div>
					<div class="border-t py-4">
						<p class="text-sm">Members</p>
						<div>
							<div class="h-30 space-y-2 overflow-scroll py-4">
								{#if req.members.length == 0}
									<ErrorMessage variant="info" text="No Members Assigned" />
								{:else}
									{#each req.members as member (member.id)}
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

							<div class="space-y-2">
								<SearchUser onSelect={addMember} />

								<div class="flex items-start gap-1 text-xs text-neutral-500">
									<Info size={16} />
									<div>
										<p>You Are Automatically Included As A Member</p>
										<p>Please Add At Least 1 Additional Member</p>
									</div>
								</div>
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
