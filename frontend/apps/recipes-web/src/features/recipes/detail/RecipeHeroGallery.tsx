import { Button, ChevronLeftIcon, ChevronRightIcon } from "@nomnomvault/ui";
import { createSignal, For, Show } from "solid-js";

type GalleryItem = {
	id: string;
	url: string;
	alt: string;
	thumbnailUrl?: string;
};

type RecipeHeroGalleryProps = {
	title: string;
	items: GalleryItem[];
};

export function RecipeHeroGallery(props: RecipeHeroGalleryProps) {
	const [selected, setSelected] = createSignal(0);
	const [pointerStartX, setPointerStartX] = createSignal<number | null>(null);

	const currentItem = () => props.items[selected()];
	const hasMultipleItems = () => props.items.length > 1;
	const blurAfterPointerClick = (
		event: MouseEvent & { currentTarget: HTMLButtonElement },
	) => {
		if (event.detail > 0) {
			event.currentTarget.blur();
		}
	};
	const goTo = (nextIndex: number) => {
		const total = props.items.length;
		if (total === 0) {
			return;
		}

		setSelected((nextIndex + total) % total);
	};

	return (
		<section
			class="grid gap-3"
			role={hasMultipleItems() ? "region" : undefined}
			tabIndex={hasMultipleItems() ? 0 : undefined}
			aria-label={
				hasMultipleItems()
					? `${props.title} carousel, image ${selected() + 1} of ${props.items.length}`
					: `${props.title} gallery`
			}
			onKeyDown={(event) => {
				if (!hasMultipleItems()) {
					return;
				}

				if (event.key === "ArrowLeft") {
					goTo(selected() - 1);
				}

				if (event.key === "ArrowRight") {
					goTo(selected() + 1);
				}
			}}
			onPointerDown={(event) => setPointerStartX(event.clientX)}
			onPointerUp={(event) => {
				const startX = pointerStartX();
				if (startX == null) {
					return;
				}

				const delta = event.clientX - startX;
				setPointerStartX(null);

				if (Math.abs(delta) < 24) {
					return;
				}

				goTo(delta > 0 ? selected() - 1 : selected() + 1);
			}}
			onPointerCancel={() => setPointerStartX(null)}
		>
			<Show
				when={currentItem()}
				fallback={
					<div class="relative grid aspect-[4/3] place-items-center overflow-hidden rounded-[var(--nnv-radius-lg)] bg-[linear-gradient(135deg,rgba(193,138,85,0.48),rgba(120,143,102,0.42))] p-4 text-[var(--nnv-text-muted)]">
						<span>No lead image is available for this recipe.</span>
					</div>
				}
			>
				<div
					class="group relative aspect-[4/3] overflow-hidden rounded-[var(--nnv-radius-lg)] bg-[linear-gradient(135deg,rgba(193,138,85,0.48),rgba(120,143,102,0.42))]"
					aria-live={hasMultipleItems() ? "polite" : undefined}
				>
					<Show when={hasMultipleItems()}>
						<Button
							variant="secondary"
							size="icon"
							class="absolute top-1/2 left-3 z-[2] h-12 w-12 -translate-y-1/2 hover:translate-y-[-50%]"
							aria-label="Previous image"
							onClick={(event) => {
								goTo(selected() - 1);
								blurAfterPointerClick(event);
							}}
						>
							<ChevronLeftIcon size="lg" />
						</Button>
					</Show>
					<img
						class="block h-full w-full bg-[var(--nnv-surface-2)] object-cover"
						src={currentItem()?.url}
						alt={currentItem()?.alt}
						loading="eager"
					/>
					<Show when={hasMultipleItems()}>
						<Button
							variant="secondary"
							size="icon"
							class="absolute top-1/2 right-3 z-[2] h-12 w-12 -translate-y-1/2 hover:translate-y-[-50%]"
							aria-label="Next image"
							onClick={(event) => {
								goTo(selected() + 1);
								blurAfterPointerClick(event);
							}}
						>
							<ChevronRightIcon size="lg" />
						</Button>
					</Show>

					<Show when={hasMultipleItems()}>
						<div class="absolute inset-x-0 bottom-0 z-[2] p-2 md:p-3">
							<div class="rounded-[calc(var(--nnv-radius-lg)-0.35rem)] bg-[color:color-mix(in_oklab,var(--nnv-surface-solid)_86%,transparent)] p-1.5 shadow-[var(--nnv-shadow-md)] backdrop-blur-[10px] transition-opacity duration-150 md:pointer-events-none md:opacity-0 md:group-hover:pointer-events-auto md:group-hover:opacity-100 md:group-focus-within:pointer-events-auto md:group-focus-within:opacity-100">
								<div class="hidden justify-center gap-1.5 md:flex">
									<For each={props.items}>
										{(item, index) => (
											<button
												type="button"
												class="h-14 w-14 overflow-hidden rounded-[var(--nnv-radius-sm)] border border-[color:var(--nnv-line)] bg-[var(--nnv-surface-1)] p-0 data-[selected=true]:shadow-[inset_0_0_0_2px_var(--nnv-accent-strong)]"
												data-selected={
													index() === selected() ? "true" : "false"
												}
												onClick={(event) => {
													setSelected(index());
													blurAfterPointerClick(event);
												}}
											>
												<img
													class="block h-full w-full object-cover"
													src={item.thumbnailUrl ?? item.url}
													alt={item.alt}
													loading="lazy"
												/>
											</button>
										)}
									</For>
								</div>
								<div class="flex flex-wrap justify-center gap-2 md:hidden">
									<For each={props.items}>
										{(_, index) => (
											<button
												type="button"
												class="h-[0.55rem] w-[0.55rem] rounded-full border-0 bg-[color:color-mix(in_oklab,var(--nnv-surface-3)_84%,var(--nnv-text-muted))] p-0 data-[selected=true]:bg-[var(--nnv-accent-strong)]"
												aria-label={`Go to image ${index() + 1}`}
												data-selected={
													index() === selected() ? "true" : "false"
												}
												onClick={(event) => {
													setSelected(index());
													blurAfterPointerClick(event);
												}}
											/>
										)}
									</For>
								</div>
							</div>
						</div>
					</Show>
				</div>
			</Show>
		</section>
	);
}
