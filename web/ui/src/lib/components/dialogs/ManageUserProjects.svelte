<script lang="ts">
	import { Dialog } from 'bits-ui';
	import X from 'phosphor-svelte/lib/X';
	import {
		getAllProjectStatsByMember,
		replaceMemberProjects,
		type ProjectStat
	} from '$lib/api/projects';
	import { toast } from 'svelte-sonner';
	import Spinner from '../Spinner.svelte';
	import ErrorMessage from '../ErrorMessage.svelte';
	import SearchProject from '../SearchProject.svelte';
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';
	import { ApiError } from '$lib/api/client';

	interface Props {
		open: boolean;
		userId: string;
		onSuccess?: () => unknown;
	}

	let { open = $bindable(), userId, onSuccess }: Props = $props();
	let confirmProjects: ProjectStat[] = $state([]);

	let loading = $state(false);
	let loadError: ApiError | null = $state(null);
	let abort: AbortController | null = null;
	async function loadProjects() {
		if (abort) {
			abort.abort();
		}
		abort = new AbortController();

		try {
			loading = true;
			confirmProjects = await getAllProjectStatsByMember(userId, abort.signal);
		} catch (error) {
			loadError = error as ApiError;
			confirmProjects = [];
		} finally {
			loading = false;
		}
	}

	function addProject(project: ProjectStat) {
		const exists = confirmProjects.find((m) => m.id == project.id);
		if (!exists) {
			confirmProjects.push(project);
			confirmProjects = [...confirmProjects];
		} else {
			toast.info('Already Added');
		}
	}

	function removeProject(project: ProjectStat) {
		const exists = confirmProjects.find((m) => m.id == project.id);
		if (exists) {
			confirmProjects = confirmProjects.filter((m) => m.id != project.id);
		}
	}

	let confirmAbort: AbortController | null = null;
	async function handleConfirm() {
		if (confirmAbort) {
			confirmAbort.abort();
		}
		confirmAbort = new AbortController();

		try {
			await replaceMemberProjects(
				userId,
				{ projectIds: confirmProjects.map((m) => m.id) },
				confirmAbort.signal
			);
			toast.success('Assigned Projects Successfully');
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
			loadProjects();
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
				<p class="font-bold">Manage Projects</p>
				<div class="mt-6 space-y-2">
					<div class="h-50 space-y-2 overflow-scroll">
						{#if loading}
							<Spinner />
						{/if}

						{#if loadError}
							<ErrorMessage variant="warn" text={loadError.message} retry={loadProjects} />
						{/if}

						{#if confirmProjects.length == 0}
							<ErrorMessage variant="info" text="No Projects Assigned" />
						{:else}
							{#each confirmProjects as project (project)}
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
					<SearchProject onSelect={(project) => addProject(project)} />
				</div>

				<div class="mt-6 space-x-1 text-right text-xs *:cursor-pointer *:rounded *:px-6 *:py-3">
					<Dialog.Close class="hover:bg-neutral-100 dark:hover:bg-neutral-900">Cancel</Dialog.Close>
					<button
						disabled={confirmProjects.length == 0}
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
