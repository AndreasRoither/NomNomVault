import { Button } from "@nomnomvault/ui";
import { For } from "solid-js";

type FilterChipRowProps = {
	title: string;
	items: { value: string; label: string }[];
	selected?: string;
	onSelect: (value?: string) => void;
};

export function FilterChipRow(props: FilterChipRowProps) {
	return (
		<section class="grid gap-2">
			<div class="text-sm font-bold text-[var(--nnv-text-muted)]">
				{props.title}
			</div>
			<ul class="m-0 flex list-none gap-2 overflow-x-auto p-0 pb-1 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
				<li>
					<Button
						variant="chip"
						size="sm"
						data-selected={props.selected ? "false" : "true"}
						onClick={() => props.onSelect(undefined)}
					>
						All
					</Button>
				</li>
				<For each={props.items}>
					{(item) => (
						<li>
							<Button
								variant="chip"
								size="sm"
								data-selected={props.selected === item.value ? "true" : "false"}
								onClick={() =>
									props.onSelect(
										props.selected === item.value ? undefined : item.value,
									)
								}
							>
								{item.label}
							</Button>
						</li>
					)}
				</For>
			</ul>
		</section>
	);
}
