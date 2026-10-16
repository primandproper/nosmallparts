import type { RoleOffer } from '../generated/nosmallparts/stage/v1/stage';

interface Props {
  offers: RoleOffer[];
  onRespond: (offerId: string, accept: boolean) => void;
}

export function RoleOfferPrompt({ offers, onRespond }: Props) {
  // Only the oldest pending offer shows; the others queue behind it.
  const offer = offers[0];
  if (!offer) return null;
  return (
    <div className="border-b border-stage-gold bg-stage-surface px-3 py-2">
      <p className="text-xs uppercase tracking-widest text-stage-gold">Role offered</p>
      <p className="mt-1 text-stage-text">
        The host has offered you <span className="font-semibold">{offer.castMemberName}</span>.
      </p>
      <p className="mt-0.5 text-xs text-stage-muted">
        Accepting puts you in the cast, so you can turn on your camera and mic to perform.
      </p>
      <div className="mt-2 flex gap-2">
        <button
          type="button"
          onClick={() => onRespond(offer.offerId, true)}
          className="rounded-md bg-stage-gold px-3 py-1 text-xs font-semibold text-stage-bg"
        >
          Accept
        </button>
        <button
          type="button"
          onClick={() => onRespond(offer.offerId, false)}
          className="rounded-md border border-stage-border px-3 py-1 text-xs text-stage-muted"
        >
          Decline
        </button>
      </div>
    </div>
  );
}
