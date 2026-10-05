<script lang="ts">
	import { formatDate } from '$lib/utils/formatDate';
	import ContractCreated, { type ContractCreatedPayload } from './ContractCreated.svelte';
	import ContractStatus, { type ContractStatusPayload } from './ContractStatus.svelte';
	import FileUploaded, { type FileUploadedPayload } from './FileUploaded.svelte';
	import InvoiceCreated, { type InvoiceCreatedPayload } from './InvoiceCreated.svelte';
	import InvoiceStatus, { type InvoiceStatusPayload } from './InvoiceStatus.svelte';
	import QuoteConverted, { type QuoteConvertedPayload } from './QuoteConverted.svelte';

	interface Props {
		message: string;
		date: Date;
	}
	let { message, date }: Props = $props();

	type SystemMessagePayload =
		| QuoteConvertedPayload
		| InvoiceStatusPayload
		| InvoiceCreatedPayload
		| FileUploadedPayload
		| ContractStatusPayload
		| ContractCreatedPayload;

	let payload: SystemMessagePayload = JSON.parse(message);
</script>

<div>
	<p class="mr-4 text-right text-xs text-neutral-500">{formatDate(date)}</p>

	{#if payload.type === 'invoice.created'}
		<InvoiceCreated message={payload} />
	{:else if payload.type === 'invoice.status.changed'}
		<InvoiceStatus message={payload} />
	{:else if payload.type === 'invoice.converted'}
		<QuoteConverted message={payload} />
	{:else if payload.type === 'contract.created'}
		<ContractCreated message={payload} />
	{:else if payload.type === 'contract.status.changed'}
		<ContractStatus message={payload} />
	{:else if payload.type === 'message.file.created'}
		<FileUploaded message={payload} />
	{/if}
</div>
