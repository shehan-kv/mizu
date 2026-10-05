<script lang="ts">
	import X from 'phosphor-svelte/lib/X';
	import InputLabel from '../InputLabel.svelte';
	import TextEditor from '../TextEditor.svelte';
	import AiSuggestionsButton from '../AiSuggestionsButton.svelte';
	import { toast } from 'svelte-sonner';
	import { createContract } from '$lib/api/contracts';
	import { createDialogState } from './createDialogState.svelte';
	import ConfirmDiscardData from './ConfirmDiscardData.svelte';
	import { ApiError, BASE_URL } from '$lib/api/client';
	import ErrorMessage from '../ErrorMessage.svelte';
	import type { ProjectMember } from '$lib/api/projects';
	import SearchProjectMember from '../SearchProjectMember.svelte';
	import FullScreenDialog from './FullScreenDialog.svelte';
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';
	import { toTitleCase } from '$lib/utils/toTitleCase';

	interface Props {
		open: boolean;
		projectId: string;
		onSuccess?: () => unknown;
	}
	let { open = $bindable(), projectId, onSuccess }: Props = $props();

	let req = $state<{
		name: string;
		terms: string;
		signatories: ProjectMember[];
	}>({
		name: '',
		terms: '',
		signatories: []
	});

	function validateReq() {
		return (
			req.name.trim().length != 0 && req.terms.trim().length != 0 && req.signatories.length != 0
		);
	}
	function resetReq() {
		req.name = '';
		req.terms = '';
		req.signatories = [];
	}

	function addSignatory(member: ProjectMember) {
		const exists = req.signatories.find((m) => m.id == member.id);
		if (!exists) {
			req.signatories.push(member);
			req.signatories = [...req.signatories];
		} else {
			toast.info('Already Added');
		}
	}

	function removeSignatory(member: ProjectMember) {
		req.signatories = req.signatories.filter((m) => m.id != member.id);
	}

	let createAbort: AbortController | null = $state(null);
	let isCreating = $state(false);

	async function handleCreate() {
		if (!validateReq()) {
			toast.error('Missing Required Fields');
			return;
		}

		if (req.terms.length < 50) {
			toast.error('Terms Must Be At Least 50 Characters Long');
		}

		if (createAbort) {
			createAbort.abort();
		}
		createAbort = new AbortController();

		try {
			isCreating = true;
			await createContract(
				projectId,
				{
					name: req.name,
					terms: req.terms,
					signatoryIds: req.signatories.map((s) => s.id)
				},
				createAbort.signal
			);
			toast.success('Successfully Created');
			resetReq();
			onSuccess?.();
			open = false;
		} catch (error) {
			if (error instanceof ApiError) {
				toast.error(toTitleCase(error.message));
			} else {
				toast.error('An Error Occurred');
			}
		} finally {
			isCreating = false;
		}
	}

	let discardDialog = createDialogState();

	function handleDiscard() {
		resetReq();
		discardDialog.close();
		open = false;
	}
</script>

<FullScreenDialog
	bind:open
	onOpenChange={(state) => {
		if (!state && (req.name.length > 0 || req.terms.length > 0 || req.signatories.length > 0)) {
			open = true;
			discardDialog.open();
		} else {
			resetReq();
			open = false;
		}
	}}
>
	<div class="grid auto-rows-[min-content_1fr] gap-4 overflow-scroll px-5">
		<div class="flex-none">
			<div class="container mx-auto">
				<p class="font-bold">Create New Contract</p>
			</div>
		</div>

		<div class="container mx-auto grid grid-cols-4 gap-2 overflow-hidden">
			<div
				class="col-span-3 grid h-full auto-rows-[min-content_1fr] gap-2 overflow-y-auto rounded border"
			>
				<div
					class="sticky top-0 flex items-center justify-end bg-neutral-100 px-4 py-1 dark:bg-neutral-900"
				>
					<AiSuggestionsButton
						class="cursor-pointer rounded p-2 text-xs transition hover:bg-neutral-200 
						disabled:cursor-default dark:bg-neutral-950 dark:hover:bg-neutral-800"
					/>
				</div>
				<div class="overflow-y-auto p-6">
					<TextEditor
						bind:value={req.terms}
						placeholder="Write contract terms here (minimum 50 characters)"
					/>
				</div>
			</div>
			<div class="grid auto-rows-[min-content_1fr_min-content] space-y-2 overflow-y-auto">
				<div class="space-y-1 rounded border p-4 text-sm *:block">
					<InputLabel htmlFor="name" text="Contract Name" required />
					<input
						type="text"
						id="name"
						bind:value={req.name}
						class="w-full rounded border border-neutral-200 bg-neutral-100 p-1.5
							outline-hidden dark:border-neutral-800 dark:bg-neutral-900"
					/>
				</div>

				<div class="grid auto-rows-[min-content_1fr] gap-4 rounded border p-4">
					<div>
						<SearchProjectMember onSelect={(member) => addSignatory(member)} {projectId} />
					</div>

					<div class="space-y-3 overflow-scroll">
						{#if req.signatories.length > 0}
							{#each req.signatories as signatory (signatory.id)}
								<div class="flex items-center gap-1">
									<div class="flex gap-1.5">
										<div
											class="size-10 overflow-hidden rounded-full bg-neutral-200/80 dark:bg-neutral-800"
										>
											{#if signatory.hasImage}
												<img
													src={`${BASE_URL}users/profile-images/${signatory.id}`}
													alt={`${signatory.firstName} ${signatory.lastName} profile picture`}
													class="size-full object-cover"
												/>
											{:else}
												<div
													class="flex size-full items-center justify-center text-xs text-neutral-500"
												>
													{signatory.firstName[0]}
												</div>
											{/if}
										</div>
										<div>
											<p class="text-sm">{signatory.firstName} {signatory.lastName}</p>
											<p class="text-xs text-neutral-500 dark:text-neutral-400">
												{#if signatory.title || signatory.role}
													{signatory.title ? signatory.title : ''}
													{signatory.role
														? `${signatory.title ? ' - ' : ''}${toTitleCaseDashed(signatory.role)}`
														: ''}
												{:else}
													N/A
												{/if}
											</p>
										</div>
									</div>

									<div class="ml-auto">
										<button
											class="cursor-pointer rounded p-1 dark:hover:bg-neutral-800"
											onclick={() => removeSignatory(signatory)}><X /></button
										>
									</div>
								</div>
							{/each}
						{:else}
							<ErrorMessage variant="info" text="Signatories Not Added" />
						{/if}
					</div>
				</div>

				<div>
					<button
						disabled={isCreating}
						onclick={handleCreate}
						class="w-full cursor-pointer rounded bg-neutral-800 p-3
						  text-sm text-neutral-50 transition
						hover:bg-neutral-950 dark:bg-neutral-200 dark:text-neutral-950
						 hover:dark:bg-neutral-50"
					>
						{#if isCreating}
							Creating...
						{:else}
							Create
						{/if}
					</button>
				</div>
			</div>
		</div>
	</div>
</FullScreenDialog>

<ConfirmDiscardData open={discardDialog.isOpen} onDiscard={handleDiscard} />
