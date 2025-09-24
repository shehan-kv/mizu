<script lang="ts">
	interface Props {
		label: string;
		type: 'text' | 'number';
		value: string | number;
		onchange: () => any;
		id: string;
		min?: number;
		max?: number;
	}
	let { label, type, value = $bindable(), onchange, id, min, max }: Props = $props();
</script>

<div class="flex h-full items-center gap-3 rounded bg-neutral-100 pl-2 pr-1 dark:bg-neutral-900">
	<label for={id} class="text-xs">{label}</label>
	<input
		class="max-w-20 rounded bg-white px-2 py-1 text-xs [appearance:textfield] dark:bg-neutral-950 [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
		name={id}
		{id}
		{type}
		{max}
		{min}
		bind:value
		{onchange}
	/>
</div>
