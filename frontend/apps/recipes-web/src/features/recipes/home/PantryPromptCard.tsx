import {
	Button,
	Card,
	CardContent,
	CardFooter,
	CardHeader,
	CardTitle,
} from "@nomnomvault/ui";

export function PantryPromptCard() {
	return (
		<Card>
			<CardHeader>
				<p class="nnv-eyebrow">Cook from your pantry</p>
				<CardTitle>Pantry matching is staged next</CardTitle>
			</CardHeader>
			<CardContent>
				<p class="m-0 text-[var(--nnv-text-muted)]">
					The pantry service is not connected yet, but this slot is fixed in the
					product layout so strong matches and almost-there recipes can land
					here without reshuffling the page.
				</p>
			</CardContent>
			<CardFooter class="flex-wrap">
				<Button variant="primary">Manage pantry later</Button>
				<Button variant="secondary">View planned matches</Button>
			</CardFooter>
		</Card>
	);
}
