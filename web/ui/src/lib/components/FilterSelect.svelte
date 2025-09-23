<script lang="ts">
	import { Select } from 'bits-ui';
	import Check from 'phosphor-svelte/lib/Check';
	import CaretUpDown from 'phosphor-svelte/lib/CaretUpDown';
	import CaretDoubleUp from 'phosphor-svelte/lib/CaretDoubleUp';
	import CaretDoubleDown from 'phosphor-svelte/lib/CaretDoubleDown';

	interface Props {
		value: string;
		onchange: () => any;
		options: { value: string; label: string }[];
		name: string;
	}
	// providing an API similar to the native
	// HTML Select with bindable value and onchange props
	let { value = $bindable(), onchange, options, name }: Props = $props();

	// let inputValue = $state<string>('');
	const selectedLabel = $derived(
		value ? options.find((option) => option.value === value)?.label : 'All'
	);
</script>

<div class="flex h-full items-center gap-3 rounded bg-neutral-100 pl-2 pr-1 dark:bg-neutral-900">
	<p class="text-xs">{name}</p>
	<Select.Root
		type="single"
		onValueChange={(v) => {
			value = v;
			onchange();
		}}
		items={options}
		allowDeselect={true}
	>
		<Select.Trigger
			class="inline-flex w-32 items-center gap-2 rounded bg-white px-2 py-1 text-xs dark:bg-neutral-950"
			aria-label="Select a theme"
		>
			{selectedLabel}
			<CaretUpDown class="text-muted-foreground ml-auto size-3" />
		</Select.Trigger>
		<Select.Portal>
			<Select.Content
				class="focus-override border-muted bg-background shadow-popover data-[state=open]:animate-in 
            data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 
            data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 
            data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 
            data-[side=top]:slide-in-from-bottom-2 outline-hidden z-50 max-h-[var(--bits-select-content-available-height)] 
            w-fit min-w-[var(--bits-select-anchor-width)] select-none rounded-xl border px-1 py-3 
            data-[side=bottom]:translate-y-1 data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 
            data-[side=top]:-translate-y-1"
				sideOffset={10}
			>
				<Select.ScrollUpButton class="flex w-full items-center justify-center">
					<CaretDoubleUp class="size-3" />
				</Select.ScrollUpButton>
				<Select.Viewport class="p-1">
					{#each options as option, i (i + option.value)}
						<Select.Item
							class="data-highlighted:bg-muted outline-hidden data-disabled:opacity-50 flex h-fit 
                        w-full select-none items-center gap-1 rounded px-4 py-2 text-xs capitalize"
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
