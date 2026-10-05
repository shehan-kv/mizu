<script lang="ts">
	import PaperPlaneRight from 'phosphor-svelte/lib/PaperPlaneRight';

	let { onclick, disabled = false } = $props();
</script>

<button
	{disabled}
	{onclick}
	class="ml-2 cursor-pointer self-end rounded bg-neutral-950 p-2 text-neutral-50
    transition disabled:bg-neutral-50 disabled:text-neutral-500 dark:bg-neutral-50
	dark:text-neutral-950 dark:disabled:bg-neutral-950 dark:disabled:text-neutral-400
	"
>
	<PaperPlaneRight size={16} />
</button>
