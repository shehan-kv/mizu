<script lang="ts">
	import { currencyFormatter } from '$lib/utils/currencyFormatter';
	import Invoice from 'phosphor-svelte/lib/Invoice';

	export interface QuoteConvertedPayload {
		type: 'invoice.converted';
		documentType: string;
		invoiceId: string;
		currencyCode: string;
		subTotal: Intl.StringNumericLiteral;
	}

	interface Props {
		message: QuoteConvertedPayload;
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
		<Invoice size={16} />
		<p>Quote Converted to Invoice</p>
	</div>

	<div class="mt-3">
		<p class="text-3xl">
			{currencyFormatter(message.currencyCode, message.subTotal)}
			<span class="text-xs">{message.currencyCode.toUpperCase()}</span>
		</p>
	</div>
</div>
