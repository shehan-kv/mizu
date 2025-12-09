<script lang="ts">
	import * as Table from '$lib/components/ui/table';
	import { getFilesByProject, type File } from '$lib/api/files';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import { onMount } from 'svelte';
	import Spinner from './Spinner.svelte';
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';
	import ErrorMessage from './ErrorMessage.svelte';
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';

	interface Props {
		projectId: number;
	}
	let { projectId }: Props = $props();

	let filesPromise: Promise<PaginatedResponse<File>> | null = $state(null);
	let abort: AbortController | null = null;
	function loadFiles() {
		if (abort) {
			abort.abort();
		}
		abort = new AbortController();

		filesPromise = getFilesByProject(projectId, '', 1, 20, abort.signal);
	}

	onMount(() => {
		loadFiles();
	});
</script>

<div
	class="grid h-full w-full grid-rows-[min-content_1fr]
	overflow-hidden rounded border"
>
	<div class="flex items-center justify-between bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
		<p class="text-sm">Files</p>
		<a href={`/projects/${projectId}/files`} class="flex items-center gap-1 text-sm">
			<span>View All</span>
			<ArrowRight />
		</a>
	</div>
	<div class="overflow-scroll px-6 py-2">
		{#await filesPromise}
			<Spinner />
		{:then res}
			{#if res && res.data.length > 0}
				<Table.Root>
					<Table.Body>
						{#each res.data as file}
							<Table.Row
								class="text-neutral-600 hover:bg-transparent hover:text-neutral-950 
	        						dark:text-neutral-400 dark:hover:text-neutral-50"
							>
								<Table.Cell class="pl-0">
									{file.originalName}
								</Table.Cell>
								<Table.Cell>
									{file.user.firstName}
									{file.user.lastName}
								</Table.Cell>
								<Table.Cell class="pr-0" align="right">
									<a href={file.url} class="inline-block cursor-pointer">
										<DownloadSimple size={18} />
									</a>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			{:else}
				<ErrorMessage variant="warn" text="Files Not Found" />
			{/if}
		{:catch err}
			{#if err instanceof APIBadRequestError}
				<ErrorMessage variant="warn" text="Invalid Request" retry={loadFiles} />
			{:else if err instanceof APIForbiddenError}
				<ErrorMessage
					variant="warn"
					text="You Don't Have Permission To View Files"
					retry={loadFiles}
				/>
			{:else if err instanceof APINotFoundError}
				<ErrorMessage variant="info" text="Not Found" retry={loadFiles} />
			{:else if err instanceof APIServerError}
				<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadFiles} />
			{:else}
				<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadFiles} />
			{/if}
		{/await}
	</div>
</div>
