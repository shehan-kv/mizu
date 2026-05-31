<script lang="ts">
	import { Label, RadioGroup, Select } from 'bits-ui';
	import { Decimal } from 'decimal.js';
	import * as Table from '$lib/components/ui/table';
	import FullScreenDialog from './FullScreenDialog.svelte';
	import Check from 'phosphor-svelte/lib/Check';
	import CaretDoubleDown from 'phosphor-svelte/lib/CaretDoubleDown';
	import CaretDoubleUp from 'phosphor-svelte/lib/CaretDoubleUp';
	import CaretUpDown from 'phosphor-svelte/lib/CaretUpDown';
	import { currencies } from '$lib/utils/currencies';
	import { currencyFormatter } from '$lib/utils/currencyFormatter';
	import DecimalInput from '../DecimalInput.svelte';
	import X from 'phosphor-svelte/lib/X';
	import { toast } from 'svelte-sonner';
	import Info from 'phosphor-svelte/lib/Info';
	import { createDialogState } from './createDialogState.svelte';
	import { createInvoice } from '$lib/api/invoices';
	import ConfirmDiscardData from './ConfirmDiscardData.svelte';
	import { ApiError } from '$lib/api/client';

	interface Props {
		open: boolean;
		projectId: string;
		onSuccess?: () => unknown;
	}

	let { open = $bindable(), projectId, onSuccess }: Props = $props();

	let req = $state<{
		type: 'invoice' | 'quote';
		currency: string;
		note: string;
		items: {
			description: string;
			qty: Decimal;
			unitPrice: Decimal;
			unitTax: Decimal;
			taxType: 'fixed' | 'percentage';
			unitDiscount: Decimal;
			discountType: 'fixed' | 'percentage';
			totalDiscount: Decimal;
			totalTax: Decimal;
			total: Decimal;
		}[];
	}>({
		type: 'invoice',
		currency: 'USD',
		note: '',
		items: []
	});

	function resetReq() {
		req.type = 'invoice';
		req.currency = 'USD';
		req.note = '';
		req.items = [];
	}

	let currencyList = $state(
		currencies.map((c) => ({ value: c.code, label: `${c.name} - ${c.symbol}` }))
	);

	let selectedCurrencySymbol = $derived(
		currencies.find((c) => c.code == req.currency)?.symbol || '.'
	);

	let discountTypes = $derived([
		{ value: 'percentage', label: '%' },
		{ value: 'fixed', label: selectedCurrencySymbol }
	]);

	let taxTypes = $derived([
		{ value: 'percentage', label: '%' },
		{ value: 'fixed', label: selectedCurrencySymbol }
	]);

	let itemDescription = $state('');
	let itemQty = $state(new Decimal(0));
	let itemUnitPrice = $state(new Decimal(0));
	let itemUnitDiscount = $state(new Decimal(0));
	let itemUnitTax = $state(new Decimal(0));
	let itemDiscountType: 'fixed' | 'percentage' = $state('percentage');
	let itemTaxType: 'fixed' | 'percentage' = $state('percentage');

	// svelte-ignore non_reactive_update
	let itemQtyInput: DecimalInput;
	// svelte-ignore non_reactive_update
	let itemUnitPriceInput: DecimalInput;
	// svelte-ignore non_reactive_update
	let itemUnitDiscountInput: DecimalInput;
	// svelte-ignore non_reactive_update
	let itemUnitTaxInput: DecimalInput;

	function resetItem() {
		itemDescription = '';
		itemDiscountType = 'percentage';
		itemTaxType = 'percentage';

		itemQtyInput.clear();
		itemUnitPriceInput.clear();
		itemUnitDiscountInput.clear();
		itemUnitTaxInput.clear();
	}

	let invTotalDiscount = $derived(
		req.items.reduce((sum, item) => {
			return sum.add(item.totalDiscount);
		}, new Decimal(0))
	);

	let invTotalTax = $derived(
		req.items.reduce((sum, item) => {
			return sum.add(item.totalTax);
		}, new Decimal(0))
	);

	let invTotal = $derived(
		req.items.reduce((sum, item) => {
			return sum.add(item.total);
		}, new Decimal(0))
	);

	function validateItem() {
		if (!itemDescription) {
			return false;
		}

		if (!(itemDiscountType == 'fixed' || itemDiscountType == 'percentage')) {
			return false;
		}

		if (!(itemTaxType == 'fixed' || itemTaxType == 'percentage')) {
			return false;
		}

		if (
			itemQty.isNegative() ||
			itemUnitPrice.isNegative() ||
			itemUnitDiscount.isNegative() ||
			itemUnitTax.isNegative()
		) {
			return false;
		}

		return true;
	}

	function handleItemAdd() {
		if (!validateItem()) {
			toast.error('Invalid Item Data');
			return;
		}

		let effectiveUnitDiscount;
		if (itemDiscountType == 'fixed') {
			effectiveUnitDiscount = itemUnitDiscount;
		} else {
			effectiveUnitDiscount = itemUnitDiscount.dividedBy(100).times(itemUnitPrice);
		}

		let totalDiscount = effectiveUnitDiscount.times(itemQty);
		let priceAfterDiscount = itemUnitPrice.sub(effectiveUnitDiscount);

		let effectiveUnitTax;
		if (itemTaxType == 'fixed') {
			effectiveUnitTax = itemUnitTax;
		} else {
			effectiveUnitTax = itemUnitTax.dividedBy(100).times(priceAfterDiscount);
		}

		let totalTax = effectiveUnitTax.times(itemQty);
		let total = itemUnitPrice.times(itemQty).sub(totalDiscount).add(totalTax);

		let item = {
			description: itemDescription,
			qty: itemQty,
			unitPrice: itemUnitPrice,
			unitTax: itemUnitTax,
			taxType: itemTaxType,
			unitDiscount: itemUnitDiscount,
			discountType: itemDiscountType,
			totalDiscount: totalDiscount,
			totalTax: totalTax,
			total: total
		};

		req.items = [...req.items, { ...item }];

		resetItem();
	}

	function removeItem(i: number) {
		req.items = req.items.filter((_, idx) => idx != i);
	}

	let createAbort: AbortController | null = $state(null);
	async function handleCreate() {
		if (createAbort) {
			createAbort.abort();
		}
		createAbort = new AbortController();

		try {
			await createInvoice(
				projectId,
				{
					isInvoice: req.type == 'invoice' ? true : false,
					currencyCode: req.currency,
					note: req.note,
					items: req.items.map((item) => ({
						description: item.description,
						qty: item.qty.toString(),
						discountRate: item.unitDiscount.toString(),
						discountType: item.discountType,
						taxRate: item.unitTax.toString(),
						taxType: item.taxType,
						unitPrice: item.unitPrice.toString()
					}))
				},
				createAbort.signal
			);

			toast.success('Successfully Created');
			resetReq();
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
		if (!state && req.items.length > 0) {
			open = true;
			discardDialog.open();
		} else {
			resetReq();
			open = false;
		}
	}}
>
	<div class="grid auto-rows-[min-content_min-content_1fr_min-content] gap-4 overflow-scroll px-5">
		<div class="flex-none">
			<div class="container mx-auto">
				<p class="font-bold">Create New Invoice / Quote</p>
			</div>
		</div>

		<div class="container mx-auto">
			<RadioGroup.Root class="flex gap-6 text-sm font-medium" bind:value={req.type}>
				<div class="text-foreground group flex items-center transition-all select-none">
					<RadioGroup.Item
						id="invoice"
						value="invoice"
						class="border-border-input bg-background hover:border-dark-40 data-[state=checked]:border-foreground size-5 shrink-0 cursor-default rounded-full border transition-all duration-100 ease-in-out data-[state=checked]:border-6"
					/>
					<Label.Root for="invoice" class="pl-3">Invoice</Label.Root>
				</div>
				<div class="text-foreground group flex items-center transition-all select-none">
					<RadioGroup.Item
						id="quote"
						value="quote"
						class="border-border-input bg-background hover:border-dark-40 data-[state=checked]:border-foreground size-5 shrink-0 cursor-default rounded-full border transition-all duration-100 ease-in-out data-[state=checked]:border-6"
					/>
					<Label.Root for="quote" class="pl-3">Quote</Label.Root>
				</div>
			</RadioGroup.Root>

			<div class="mt-6 flex gap-4 *:space-y-1">
				<div>
					<p class="text-sm">Currency</p>
					<Select.Root
						type="single"
						bind:value={req.currency}
						items={currencyList}
						allowDeselect={false}
					>
						<Select.Trigger
							class="inline-flex items-center gap-2 rounded bg-white px-4 py-2.5 text-xs dark:bg-neutral-900"
							aria-label="Select currency"
						>
							{currencyList.find((c) => c.value == req.currency)?.label}
							<CaretUpDown class="text-muted-foreground ml-auto size-3" />
						</Select.Trigger>
						<Select.Portal>
							<Select.Content
								class="focus-override border-muted bg-background shadow-popover data-[state=open]:animate-in 
									   data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 
									data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 
									data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 
									data-[side=top]:slide-in-from-bottom-2 z-50 max-h-100 w-fit
									min-w-(--bits-select-anchor-width) rounded-xl border px-1 py-3 outline-hidden select-none 
									data-[side=bottom]:translate-y-1 data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 
									data-[side=top]:-translate-y-1"
								sideOffset={10}
							>
								<Select.ScrollUpButton class="flex w-full items-center justify-center">
									<CaretDoubleUp class="size-3" />
								</Select.ScrollUpButton>
								<Select.Viewport class="p-1">
									{#each currencyList as option, i (i + option.value)}
										<Select.Item
											class="data-highlighted:bg-muted flex h-fit w-full items-center 
												gap-1 rounded px-4 py-2 text-xs capitalize outline-hidden select-none data-disabled:opacity-50"
											value={option.value}
											label={option.label}
										>
											{#snippet children({ selected })}
												{option.label}
												{#if selected}
													<div class="ml-auto">
														<Check aria-label="check" />
													</div>
												{/if}
											{/snippet}
										</Select.Item>
									{/each}
								</Select.Viewport>
								<Select.ScrollDownButton class="flex w-full items-center justify-center">
									<CaretDoubleDown class="size-3" />
								</Select.ScrollDownButton>
							</Select.Content>
						</Select.Portal>
					</Select.Root>
				</div>
			</div>
		</div>

		<div class="container mx-auto">
			<div>
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head class="font-bold">Description</Table.Head>
							<Table.Head class="font-bold">QTY</Table.Head>
							<Table.Head class="font-bold">Unit Price</Table.Head>
							<Table.Head class="font-bold">Unit Discount</Table.Head>
							<Table.Head class="font-bold">Unit Tax</Table.Head>
							<Table.Head class="font-bold">Total Discount</Table.Head>
							<Table.Head class="font-bold">Total Tax</Table.Head>
							<Table.Head class="text-right font-bold">Item Total</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#if req.items.length > 0}
							{#each req.items as item, i (i)}
								<Table.Row>
									<Table.Cell class="flex max-w-xs items-center gap-1">
										<button
											onclick={() => removeItem(i)}
											class="cursor-pointer rounded p-2 hover:bg-neutral-800"
										>
											<X />
										</button>
										<p class="overflow-hidden text-wrap text-ellipsis">
											{item.description}
										</p>
									</Table.Cell>
									<Table.Cell>{item.qty.toString()}</Table.Cell>
									<Table.Cell>
										{currencyFormatter(
											req.currency,
											item.unitPrice.toString() as Intl.StringNumericLiteral
										)}
									</Table.Cell>
									<Table.Cell>
										{#if item.discountType == 'percentage'}
											{item.unitDiscount}%
										{:else}
											{currencyFormatter(
												req.currency,
												item.unitDiscount.toString() as Intl.StringNumericLiteral
											)}
										{/if}
									</Table.Cell>
									<Table.Cell>
										{#if item.taxType == 'percentage'}
											{item.unitTax}%
										{:else}
											{currencyFormatter(
												req.currency,
												item.unitTax.toString() as Intl.StringNumericLiteral
											)}
										{/if}
									</Table.Cell>
									<Table.Cell>
										{currencyFormatter(
											req.currency,
											item.totalDiscount.toString() as Intl.StringNumericLiteral
										)}
									</Table.Cell>
									<Table.Cell>
										{currencyFormatter(
											req.currency,
											item.totalTax.toString() as Intl.StringNumericLiteral
										)}
									</Table.Cell>
									<Table.Cell align="right">
										{currencyFormatter(
											req.currency,
											item.total.toString() as Intl.StringNumericLiteral
										)}
									</Table.Cell>
								</Table.Row>
							{/each}
						{/if}

						<Table.Row class="bg-neutral-100 hover:bg-neutral-100 dark:bg-neutral-900">
							<Table.Cell>
								<input
									bind:value={itemDescription}
									required
									placeholder="Add Description"
									type="text"
									name="description"
									id="description"
									class="rounded bg-white px-4 py-2 placeholder:text-xs placeholder:italic dark:bg-neutral-950"
								/>
							</Table.Cell>
							<Table.Cell>
								{#key req.currency}
									<DecimalInput
										bind:this={itemQtyInput}
										bind:value={itemQty}
										defaultValue="0"
										required
										placeholder="Add QTY"
										name="qty"
										id="qty"
										class="text-right"
									/>
								{/key}
							</Table.Cell>
							<Table.Cell>
								{#key req.currency}
									<DecimalInput
										bind:this={itemUnitPriceInput}
										bind:value={itemUnitPrice}
										defaultValue="0"
										required
										placeholder="Add Unit Price"
										name="unitprice"
										id="unitprice"
										class="text-right"
										currency={req.currency}
									/>
								{/key}
							</Table.Cell>
							<Table.Cell>
								<div class="flex">
									<Select.Root
										type="single"
										bind:value={itemDiscountType}
										items={discountTypes}
										allowDeselect={false}
									>
										<Select.Trigger
											class="inline-flex items-center gap-2 rounded-l bg-white px-3 py-2.5 text-xs dark:bg-neutral-950"
											aria-label="Select discount type"
										>
											<span class="w-3">
												{discountTypes.find((t) => t.value == itemDiscountType)?.label}
											</span>
											<CaretUpDown class="text-muted-foreground ml-auto size-3" />
										</Select.Trigger>
										<Select.Portal>
											<Select.Content
												class="focus-override border-muted bg-background shadow-popover data-[state=open]:animate-in 
									   				data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 
													data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 
													data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 
													data-[side=top]:slide-in-from-bottom-2 z-50 max-h-100 w-fit
													min-w-(--bits-select-anchor-width) rounded-xl border px-1 py-3 outline-hidden select-none 
													data-[side=bottom]:translate-y-1 data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 
													data-[side=top]:-translate-y-1"
												sideOffset={10}
											>
												<Select.Viewport class="p-1">
													{#each discountTypes as option, i (i + option.value)}
														<Select.Item
															class="data-highlighted:bg-muted flex h-fit w-full items-center 
																gap-1 rounded px-4 py-2 text-xs capitalize outline-hidden select-none data-disabled:opacity-50"
															value={option.value}
															label={option.label}
														>
															{#snippet children({ selected })}
																{option.label}
																{#if selected}
																	<div class="ml-auto">
																		<Check aria-label="check" />
																	</div>
																{/if}
															{/snippet}
														</Select.Item>
													{/each}
												</Select.Viewport>
												<Select.ScrollDownButton class="flex w-full items-center justify-center">
													<CaretDoubleDown class="size-3" />
												</Select.ScrollDownButton>
											</Select.Content>
										</Select.Portal>
									</Select.Root>

									{#key (req.currency, itemDiscountType)}
										<DecimalInput
											bind:this={itemUnitDiscountInput}
											bind:value={itemUnitDiscount}
											defaultValue="0"
											required
											placeholder="Add Unit Discount"
											name="unitdiscount"
											id="unitdiscount"
											class="w-36 rounded-none rounded-r text-right"
											currency={itemDiscountType == 'fixed' ? req.currency : undefined}
										/>
									{/key}
								</div>
							</Table.Cell>
							<Table.Cell>
								<div class="flex">
									<Select.Root
										type="single"
										bind:value={itemTaxType}
										items={taxTypes}
										allowDeselect={false}
									>
										<Select.Trigger
											class="inline-flex items-center gap-2 rounded-l bg-white px-4 py-2.5 text-xs dark:bg-neutral-950"
											aria-label="Select tax type"
										>
											<span class="w-3">
												{taxTypes.find((t) => t.value == itemTaxType)?.label}
											</span>
											<CaretUpDown class="text-muted-foreground ml-auto size-3" />
										</Select.Trigger>
										<Select.Portal>
											<Select.Content
												class="focus-override border-muted bg-background shadow-popover data-[state=open]:animate-in 
									   				data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 
													data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 
													data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 
													data-[side=top]:slide-in-from-bottom-2 z-50 max-h-100 w-fit
													min-w-(--bits-select-anchor-width) rounded-xl border px-1 py-3 outline-hidden select-none 
													data-[side=bottom]:translate-y-1 data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 
													data-[side=top]:-translate-y-1"
												sideOffset={10}
											>
												<Select.Viewport class="p-1">
													{#each taxTypes as option, i (i + option.value)}
														<Select.Item
															class="data-highlighted:bg-muted flex h-fit w-full items-center 
																gap-1 rounded px-4 py-2 text-xs capitalize outline-hidden select-none data-disabled:opacity-50"
															value={option.value}
															label={option.label}
														>
															{#snippet children({ selected })}
																{option.label}
																{#if selected}
																	<div class="ml-auto">
																		<Check aria-label="check" />
																	</div>
																{/if}
															{/snippet}
														</Select.Item>
													{/each}
												</Select.Viewport>
												<Select.ScrollDownButton class="flex w-full items-center justify-center">
													<CaretDoubleDown class="size-3" />
												</Select.ScrollDownButton>
											</Select.Content>
										</Select.Portal>
									</Select.Root>

									{#key (req.currency, itemTaxType)}
										<DecimalInput
											bind:this={itemUnitTaxInput}
											bind:value={itemUnitTax}
											defaultValue="0"
											required
											placeholder="Add Unit Tax"
											name="unittax"
											id="unittax"
											class="w-36 rounded-none rounded-r text-right"
											currency={itemTaxType == 'fixed' ? req.currency : undefined}
										/>
									{/key}
								</div>
							</Table.Cell>
							<Table.Cell colspan={3}>
								<button
									onclick={handleItemAdd}
									class="cursor-pointer rounded bg-neutral-800 px-4 py-2 text-xs text-neutral-50
										hover:bg-neutral-950 dark:bg-neutral-200 dark:text-neutral-950 dark:hover:bg-neutral-50"
								>
									Add Item
								</button>
							</Table.Cell>
						</Table.Row>

						<Table.Row>
							<Table.Cell colspan={7} align="right" class="border-r">Total Discount</Table.Cell>
							<Table.Cell align="right">
								{currencyFormatter(
									req.currency,
									invTotalDiscount.toString() as Intl.StringNumericLiteral
								)}
							</Table.Cell>
						</Table.Row>
						<Table.Row>
							<Table.Cell colspan={7} align="right" class="border-r">Total Tax</Table.Cell>
							<Table.Cell align="right">
								{currencyFormatter(
									req.currency,
									invTotalTax.toString() as Intl.StringNumericLiteral
								)}
							</Table.Cell>
						</Table.Row>
						<Table.Row class="bg-neutral-100 dark:bg-neutral-900">
							<Table.Cell colspan={7} align="right" class="border-r">Invoice Total</Table.Cell>
							<Table.Cell align="right">
								{currencyFormatter(req.currency, invTotal.toString() as Intl.StringNumericLiteral)}
							</Table.Cell>
						</Table.Row>
					</Table.Body>
				</Table.Root>
			</div>

			<div class="mt-6 space-y-1 text-sm">
				<p>Note (Optional)</p>
				<textarea
					bind:value={req.note}
					name="note"
					id="note"
					class="h-40 w-full resize-none rounded bg-neutral-100 p-4 dark:bg-neutral-900"
				></textarea>
			</div>
		</div>

		<div class="container mx-auto space-y-6">
			<div class="ml-auto flex w-fit gap-1 text-neutral-600 dark:text-neutral-400">
				<Info size={18} class="mt-0.5" weight="fill" />
				<p class="max-w-sm text-sm italic">
					Once an invoice / quote is added and finalized, it cannot be altered or edited. Please
					ensure all details are correct before submission.
				</p>
			</div>

			<div class="space-x-2 text-right text-xs *:cursor-pointer *:rounded *:px-6 *:py-3">
				<button
					class="bg-neutral-100 hover:bg-neutral-200 dark:bg-neutral-900 dark:hover:bg-neutral-800"
				>
					Cancel
				</button>
				<button
					onclick={handleCreate}
					class="bg-neutral-800 text-neutral-50 hover:bg-neutral-950 dark:bg-neutral-200
				dark:text-neutral-950 dark:hover:bg-neutral-50"
				>
					Create
				</button>
			</div>
		</div>
	</div>
</FullScreenDialog>

<ConfirmDiscardData open={discardDialog.isOpen} onDiscard={handleDiscard} />
