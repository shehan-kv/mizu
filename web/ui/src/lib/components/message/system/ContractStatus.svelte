<script lang="ts">
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';
	import FileText from 'phosphor-svelte/lib/FileText';

	export interface ContractStatusPayload {
		type: 'contract.status.changed';
		status: string;
		contractId: string;
		name: string;
		userId: string;
		firstName: string;
		lastName: string;
		occurredAt: string;
	}

	interface Props {
		message: ContractStatusPayload;
	}

	// eslint-disable-next-line svelte/no-unused-props
	let { message }: Props = $props();
</script>

<div
	class="my-1 ml-auto w-fit min-w-xs gap-4 rounded
		bg-neutral-100 p-4 text-sm text-neutral-700 dark:bg-neutral-900
		dark:text-neutral-300"
>
	<p class="text-xs text-neutral-500">
		#{message.contractId.replaceAll('-', '').slice(-8).toUpperCase()}
	</p>
	<div class="flex items-center gap-1">
		<FileText size={16} weight="fill" />
		<p class="grow">Contract {toTitleCase(message.status)}</p>
		<button
			title="Download"
			class="cursor-pointer p-1 text-neutral-500 hover:text-neutral-950 dark:hover:text-neutral-50"
		>
			<DownloadSimple size={16} />
		</button>
	</div>

	<div class="mt-3">
		<p>{message.name}</p>
	</div>

	<div class="mt-3">
		<p
			class="w-fit rounded bg-neutral-200
				px-3 py-1.5 text-xs dark:bg-neutral-800"
		>
			{message.firstName}
			{message.lastName}
		</p>
	</div>
</div>
