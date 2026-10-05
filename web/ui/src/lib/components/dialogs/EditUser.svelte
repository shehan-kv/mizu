<script lang="ts">
	import X from 'phosphor-svelte/lib/X';
	import { Dialog, Select } from 'bits-ui';
	import { getUser, updateUser, type UserRole } from '$lib/api/users';
	import Spinner from '../Spinner.svelte';
	import ErrorMessage from '../ErrorMessage.svelte';
	import InputLabel from '../InputLabel.svelte';
	import CaretUpDown from 'phosphor-svelte/lib/CaretUpDown';
	import CaretDoubleUp from 'phosphor-svelte/lib/CaretDoubleUp';
	import Check from 'phosphor-svelte/lib/Check';
	import CaretDoubleDown from 'phosphor-svelte/lib/CaretDoubleDown';
	import { ApiError, BASE_URL } from '$lib/api/client';
	import { toast } from 'svelte-sonner';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import CircleNotch from 'phosphor-svelte/lib/CircleNotch';

	interface Props {
		userId: string;
		open: boolean;
		onSuccess?: () => unknown;
	}

	let { open = $bindable(), userId, onSuccess }: Props = $props();

	const roles: { value: UserRole; label: string }[] = [
		{ value: 'administrator', label: 'Administrator' },
		{ value: 'staff', label: 'Staff' },
		{ value: 'client', label: 'Client' }
	];

	let userState = $state<{
		firstName: string;
		lastName: string;
		email: string;
		title: string;
		role: string;
		hasImage: boolean;
	} | null>(null);

	let isUserLoading = $state(false);
	let loadingError: string | null = $state(null);
	let loadAbort: AbortController | null = null;

	async function loadUser() {
		isUserLoading = true;
		loadingError = null;

		loadAbort?.abort();
		loadAbort = new AbortController();

		try {
			let user = await getUser(userId, loadAbort.signal);
			userState = {
				firstName: user.firstName,
				lastName: user.lastName,
				email: user.email,
				title: user.title ?? '',
				role: user.role,
				hasImage: user.hasImage
			};
		} catch (error) {
			loadingError = error instanceof Error ? error.message : String(error);
		} finally {
			isUserLoading = false;
		}
	}

	let isUpdating = $state(false);
	let updateAbort: AbortController | null = null;
	async function handleUpdate() {
		if (!userState) return;

		isUpdating = true;

		try {
			await updateUser(
				userId,
				{
					firstName: userState.firstName,
					lastName: userState.lastName,
					email: userState.email,
					role: userState.role,
					title: userState.title ?? undefined,
					image: imageToUpload ?? undefined
				},
				{
					signal: updateAbort?.signal
				}
			);

			isUpdating = false;

			toast.success('Updated Successfully');

			onSuccess?.();
			open = false;
		} catch (error) {
			if (error instanceof ApiError) {
				toast.error(toTitleCase(error.message));
			} else {
				toast.error('An Error Occurred');
			}
			isUpdating = false;
		}
	}

	let imageToUpload: File | null = $state(null);
	let previewUrl = $state<string | null>(null);

	function handleFileChange(event: Event) {
		const input = event.target as HTMLInputElement;
		const file = input.files?.[0];

		if (!file) return;

		if (!file.type.startsWith('image/')) {
			toast.error('Please Select A Valid Image File.');
			input.value = '';
			return;
		}

		imageToUpload = file;
	}

	function removeUploadedImage() {
		imageToUpload = null;
	}

	function reset() {
		userState = null;
		imageToUpload = null;
	}

	$effect(() => {
		if (!imageToUpload) {
			previewUrl = null;
			return;
		}

		const url = URL.createObjectURL(imageToUpload);
		previewUrl = url;

		return () => URL.revokeObjectURL(url);
	});

	$effect(() => {
		if (open) {
			loadUser();
		} else {
			loadAbort?.abort();
			reset();
		}
	});
</script>

<Dialog.Root bind:open>
	<Dialog.Portal>
		<Dialog.Overlay
			class="data-[state=open]:animate-in data-[state=closed]:animate-out 
			data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed 
			inset-0 z-50 bg-neutral-100/80 dark:bg-black/80"
		/>
		<Dialog.Content
			class="bg-background data-[state=open]:animate-in data-[state=closed]:animate-out 
			data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0 
			data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95
			fixed top-1/2 left-1/2 z-50 grid w-full max-w-xl -translate-x-1/2 -translate-y-1/2 auto-rows-[min-content_1fr] gap-4 rounded outline-hidden 
			duration-250"
		>
			<div class="text-right">
				<Dialog.Close
					class="cursor-pointer rounded-bl bg-neutral-50 px-4 py-2
					transition duration-150
					hover:bg-neutral-950 hover:text-neutral-50 dark:bg-neutral-900
					hover:dark:bg-neutral-50 hover:dark:text-neutral-950"
				>
					<X class="size-3" />
				</Dialog.Close>
			</div>

			<div class="px-6 pb-6">
				{#if isUserLoading}
					<Spinner />
				{:else if loadingError}
					<ErrorMessage variant="warn" text={loadingError} retry={loadUser} />
				{:else if userState}
					<div class="space-y-8">
						<div class="flex gap-4">
							<div class="size-32 overflow-hidden rounded-full">
								{#if imageToUpload}
									<img
										src={previewUrl}
										alt={`${userState.firstName} ${userState.lastName} profile picture`}
										class="size-full object-cover"
									/>
								{:else if userState.hasImage}
									<!-- Add current timestamp to get around browser caching -->
									<img
										src={`${BASE_URL}users/profile-images/${userId}?v=${Date.now()}`}
										alt={`${userState.firstName} ${userState.lastName} profile picture`}
										class="size-full object-cover"
									/>
								{:else}
									<span
										class="flex h-full w-full items-center justify-center overflow-hidden
										bg-neutral-200 text-4xl text-neutral-600 dark:bg-neutral-800"
									>
										{userState.firstName[0].toUpperCase()}
									</span>
								{/if}
							</div>

							<div class="flex w-40 flex-col items-start justify-end gap-1">
								<label
									for="image-upload"
									class="w-full cursor-pointer rounded bg-neutral-200 px-4 py-2 text-center text-xs
									transition hover:bg-neutral-300 dark:bg-neutral-900 dark:hover:bg-neutral-800"
								>
									Upload New Image
								</label>

								<input
									id="image-upload"
									type="file"
									accept="image/*"
									class="hidden"
									onchange={handleFileChange}
								/>

								{#if imageToUpload}
									<button
										onclick={removeUploadedImage}
										class="w-full cursor-pointer rounded bg-red-500 px-4 py-2 text-xs text-red-50 transition
										hover:bg-red-600 dark:bg-red-900 dark:hover:bg-red-800"
									>
										Remove Image
									</button>
								{/if}
							</div>
						</div>
						<hr />
						<div class="grid grid-cols-2 gap-2">
							<div class="space-y-1 text-sm *:block">
								<InputLabel htmlFor="firstName" text="First Name" required />
								<input
									type="text"
									id="firstName"
									bind:value={userState.firstName}
									class="w-full rounded border border-neutral-200 bg-neutral-100 p-2
										outline-hidden dark:border-neutral-800 dark:bg-neutral-900"
								/>
							</div>
							<div class="space-y-1 text-sm *:block">
								<InputLabel htmlFor="lastName" text="Last Name" required />
								<input
									type="text"
									id="lastName"
									bind:value={userState.lastName}
									class="w-full rounded border border-neutral-200 bg-neutral-100 p-2
                        				outline-hidden dark:border-neutral-800 dark:bg-neutral-900"
								/>
							</div>
							<div class="space-y-1 text-sm *:block">
								<InputLabel htmlFor="title" text="Title" />
								<input
									type="text"
									id="title"
									bind:value={userState.title}
									class="w-full rounded border border-neutral-200 bg-neutral-100 p-2
                        				outline-hidden dark:border-neutral-800 dark:bg-neutral-900"
								/>
							</div>
							<div class="space-y-1 text-sm *:block">
								<InputLabel htmlFor="email" text="Email" required />
								<input
									type="email"
									id="email"
									bind:value={userState.email}
									class="w-full rounded border border-neutral-200 bg-neutral-100 p-2
                        				outline-hidden dark:border-neutral-800 dark:bg-neutral-900"
								/>
							</div>
							<div class="space-y-1 text-sm *:block">
								<InputLabel htmlFor="role" text="Role" />
								<Select.Root
									type="single"
									bind:value={userState.role}
									items={roles}
									allowDeselect={false}
								>
									<Select.Trigger
										class="w-32 gap-2 rounded border border-neutral-200 
                                    bg-neutral-100 px-2 py-2.5 dark:border-neutral-800 dark:bg-neutral-900"
										aria-label="Select a status"
									>
										<p class="flex items-center text-xs">
											{roles.find((role) => role.value === userState?.role)?.label}
											<CaretUpDown class="text-muted-foreground ml-auto size-3" />
										</p>
									</Select.Trigger>

									<Select.Portal>
										<Select.Content
											class="focus-override border-muted bg-background shadow-popover data-[state=open]:animate-in 
                                        data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 
                                        data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 
                                        data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 
                                        data-[side=top]:slide-in-from-bottom-2 z-50 max-h-(--bits-select-content-available-height) 
                                        w-fit min-w-(--bits-select-anchor-width) rounded-xl 
                                        border px-1 py-3 outline-hidden select-none data-[side=bottom]:translate-y-1 
                                        data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 data-[side=top]:-translate-y-1"
											sideOffset={10}
										>
											<Select.ScrollUpButton class="flex w-full items-center justify-center">
												<CaretDoubleUp class="size-3" />
											</Select.ScrollUpButton>
											<Select.Viewport class="p-1">
												{#each roles as role, i (i + role.value)}
													<Select.Item
														class="data-highlighted:bg-muted flex 
                                                        h-fit w-full items-center 
                                                        gap-1 rounded px-4 py-2 text-xs 
                                                        capitalize outline-hidden select-none data-disabled:opacity-50"
														value={role.value}
														label={role.label}
													>
														{#snippet children({ selected })}
															{role.label}
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

						<hr />

						<div class="space-x-1 text-right text-xs *:cursor-pointer *:rounded *:px-6 *:py-3">
							<Dialog.Close class="hover:bg-neutral-100 dark:hover:bg-neutral-900">
								Cancel
							</Dialog.Close>
							<button
								onclick={handleUpdate}
								disabled={isUpdating}
								class="bg-neutral-800 text-neutral-50 transition hover:bg-neutral-950
                    			dark:bg-neutral-200 dark:text-neutral-950 dark:hover:bg-neutral-50"
							>
								{#if isUpdating}
									<span class="flex items-center gap-2">
										<CircleNotch class="animate-spin" size={14} />
										Updating...
									</span>
								{:else}
									<span>Update</span>
								{/if}
							</button>
						</div>
					</div>
				{/if}
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
