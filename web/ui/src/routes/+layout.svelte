<script lang="ts">
	import { Toaster } from '$lib/components/ui/sonner/index.js';
	import '../app.css';

	let { children } = $props();
</script>

{@render children()}

<Toaster
	toastOptions={{
		descriptionClass: 'text-xs !text-neutral-700 dark:!text-neutral-300'
	}}
/>
