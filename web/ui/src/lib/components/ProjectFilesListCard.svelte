<script lang="ts">
	import * as Table from '$lib/components/ui/table';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import { onDestroy, onMount } from 'svelte';
	import Spinner from './Spinner.svelte';
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';
	import ErrorMessage from './ErrorMessage.svelte';
	import type { PaginatedResponse } from '$lib/api/page';
	import { downloadChannelFile, getProjectFiles, type ChannelFile } from '$lib/api/messages';
	import { resolve } from '$app/paths';
	import { ApiError } from '$lib/api/client';

	interface Props {
		projectId: string;
	}
	let { projectId }: Props = $props();

	let filesPromise: Promise<PaginatedResponse<ChannelFile>> | null = $state(null);
	let abort: AbortController | null = null;
	function loadFiles() {
		if (abort) {
			abort.abort();
		}
		abort = new AbortController();

		filesPromise = getProjectFiles(projectId, { page: 1, limit: 20 }, abort.signal);
	}

	onMount(() => {
		loadFiles();
	});

	onDestroy(() => {
		abort?.abort();
	});
</script>

<div class="grid h-full w-full grid-rows-[min-content_1fr] overflow-hidden">
	<div class="flex items-center justify-between border-b px-6 py-2">
		<p class="text-sm">Files</p>
		<a href={resolve(`/projects/${projectId}/files`)} class="flex items-center gap-1 text-sm">
			<span>View All</span>
			<ArrowRight />
		</a>
	</div>
	<div class="overflow-auto px-6 py-2">
		{#await filesPromise}
			<Spinner />
		{:then res}
			{#if res && res.items.length > 0}
				<Table.Root>
					<Table.Body>
						{#each res.items as file (file.id)}
							<Table.Row
								class="text-neutral-600 hover:bg-transparent hover:text-neutral-950 
	        						dark:text-neutral-400 dark:hover:text-neutral-50"
							>
								<Table.Cell class="pl-0">
									{file.name}
								</Table.Cell>
								<Table.Cell>
									{file.user.firstName}
									{file.user.lastName}
								</Table.Cell>
								<Table.Cell class="pr-0" align="right">
									<button
										title="Download"
										onclick={() => downloadChannelFile(file.id)}
										class="inline-block cursor-pointer"
									>
										<DownloadSimple size={18} />
									</button>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			{:else}
				<ErrorMessage variant="warn" text="Files Not Found" />
			{/if}
		{:catch err}
			{#if err instanceof ApiError}
				<ErrorMessage variant="warn" text={err.message} retry={loadFiles} />
			{:else}
				<ErrorMessage variant="warn" text="An Error Occurred" retry={loadFiles} />
			{/if}
		{/await}
	</div>
</div>
