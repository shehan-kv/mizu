<script lang="ts">
	import { currencyFormatter } from '$lib/utils/currencyFormatter';
	import Invoice from 'phosphor-svelte/lib/Invoice';

	export interface InvoiceCreatedPayload {
		type: 'invoice.created';
		documentType: string;
		invoiceId: string;
		currencyCode: string;
		subTotal: Intl.StringNumericLiteral;
	}

	interface Props {
		message: InvoiceCreatedPayload;
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
		<p>Invoice Created</p>
	</div>

	<div class="mt-3">
		<p class="text-3xl">
			{currencyFormatter(message.currencyCode, message.subTotal)}
			<span class="text-xs">{message.currencyCode.toUpperCase()}</span>
		</p>
	</div>

	<div class="mt-3">
		<p class="w-fit rounded bg-neutral-200 px-3 py-1.5 text-xs dark:bg-neutral-800">Pending</p>
	</div>
</div>
