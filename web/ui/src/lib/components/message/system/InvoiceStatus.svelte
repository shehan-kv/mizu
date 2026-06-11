<script lang="ts">
	import { currencyFormatter } from '$lib/utils/currencyFormatter';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';
	import Invoice from 'phosphor-svelte/lib/Invoice';

	export interface InvoiceStatusPayload {
		type: 'invoice.status.changed';
		status: string;
		documentType: string;
		invoiceId: string;
		currencyCode: string;
		subTotal: Intl.StringNumericLiteral;
	}

	interface Props {
		message: InvoiceStatusPayload;
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
		#{message.invoiceId.replaceAll('-', '').slice(-8).toUpperCase()}
	</p>
	<div class="flex items-center gap-1">
		<Invoice size={16} weight="fill" />
		<p class="grow">Invoice {toTitleCase(message.status)}</p>
		<button
			title="Download"
			class="cursor-pointer p-1 text-neutral-500 hover:text-neutral-950 dark:hover:text-neutral-50"
		>
			<DownloadSimple size={16} />
		</button>
	</div>

	<div class="mt-3">
		<p class="text-3xl">
			{currencyFormatter(message.currencyCode, message.subTotal)}
			<span class="text-xs">{message.currencyCode.toUpperCase()}</span>
		</p>
	</div>
</div>
