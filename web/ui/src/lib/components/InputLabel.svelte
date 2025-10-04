<script lang="ts">
	interface Props {
		htmlFor: string;
		text: string;
		required?: boolean;
	}
	let { htmlFor, text, required = false }: Props = $props();
</script>

<label for={htmlFor}>
	{text}
	{#if required}
		<span class="text-xs text-red-600 dark:text-red-500">(Required)</span>
	{/if}
</label>
