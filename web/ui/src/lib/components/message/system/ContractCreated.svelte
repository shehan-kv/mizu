<script lang="ts">
	import FileText from 'phosphor-svelte/lib/FileText';

	export interface ContractCreatedPayload {
		type: 'contract.created';
		contractId: string;
		name: string;
		occurredAt: string;
	}

	interface Props {
		message: ContractCreatedPayload;
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
		<p>Contract Created</p>
	</div>

	<div class="mt-3">
		<p>{message.name}</p>
	</div>

	<div class="mt-3">
		<p class="w-fit rounded bg-neutral-200 px-3 py-1.5 text-xs dark:bg-neutral-800">Pending</p>
	</div>
</div>
