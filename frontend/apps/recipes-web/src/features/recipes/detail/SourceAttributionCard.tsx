import { Card, CardContent, CardHeader, CardTitle } from "@nomnomvault/ui";

type SourceAttributionCardProps = {
	source?: {
		url?: string;
		capturedAtLabel?: string;
		versionLabel?: string;
	};
};

export function SourceAttributionCard(props: SourceAttributionCardProps) {
	return (
		<Card>
			<CardHeader>
				<p class="nnv-eyebrow">Source</p>
				<CardTitle>Attribution</CardTitle>
			</CardHeader>
			<CardContent class="grid gap-3">
				{props.source?.url ? (
					<a href={props.source.url} target="_blank" rel="noreferrer">
						{props.source.url}
					</a>
				) : (
					<span>No source URL provided.</span>
				)}
				{props.source?.capturedAtLabel ? (
					<span>{props.source.capturedAtLabel}</span>
				) : null}
				{props.source?.versionLabel ? (
					<span>{props.source.versionLabel}</span>
				) : null}
			</CardContent>
		</Card>
	);
}
