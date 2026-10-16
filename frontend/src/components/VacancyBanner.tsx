import { useEffect, useState } from 'react';

import type { RoleVacancy } from '../generated/nosmallparts/stage/v1/stage';

interface Props {
  vacancies: RoleVacancy[];
  isAudience: boolean;
  onClaim: (castMemberId: string) => void;
}

/**
 * Shown while a role is vacant mid-performance (§4.5). Outside director mode a
 * live countdown runs against the server's claimDeadline and audience members
 * get a one-click claim button. In director mode (no claimDeadline) the
 * director recasts through the offer flow.
 */
export function VacancyBanner({ vacancies, isAudience, onClaim }: Props) {
  if (vacancies.length === 0) return null;
  return (
    <div>
      {vacancies.map((v) => (
        <VacancyRow key={v.castMemberId} vacancy={v} isAudience={isAudience} onClaim={onClaim} />
      ))}
    </div>
  );
}

function VacancyRow({
  vacancy,
  isAudience,
  onClaim,
}: {
  vacancy: RoleVacancy;
  isAudience: boolean;
  onClaim: (castMemberId: string) => void;
}) {
  const deadline = vacancy.claimDeadline ? vacancy.claimDeadline.getTime() : null;
  const secondsLeft = useCountdown(deadline);
  const expired = deadline !== null && secondsLeft <= 0;

  return (
    <div className="border-b border-amber-500 bg-stage-surface px-3 py-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex-1">
          <p className="text-xs uppercase tracking-widest text-amber-400">Role vacant</p>
          <p className="mt-0.5 text-stage-text">
            <span className="font-semibold">{vacancy.castMemberName}</span>
            {deadline === null
              ? ' needs a new performer. The director will recast the role.'
              : expired
                ? ' went unclaimed.'
                : ` needs a new performer: ${secondsLeft}s left to claim.`}
          </p>
        </div>
        {isAudience && deadline !== null && !expired ? (
          <button
            type="button"
            onClick={() => onClaim(vacancy.castMemberId)}
            aria-label={`Claim the role ${vacancy.castMemberName}`}
            className="rounded-md bg-stage-gold px-3 py-1.5 text-xs font-semibold text-stage-bg active:opacity-80"
          >
            Claim {vacancy.castMemberName}
          </button>
        ) : null}
      </div>
    </div>
  );
}

function useCountdown(deadline: number | null): number {
  const [secondsLeft, setSecondsLeft] = useState(() =>
    deadline === null ? 0 : Math.max(0, Math.ceil((deadline - Date.now()) / 1000)),
  );
  useEffect(() => {
    if (deadline === null) return;
    const tick = () => setSecondsLeft(Math.max(0, Math.ceil((deadline - Date.now()) / 1000)));
    tick();
    const t = setInterval(tick, 250);
    return () => clearInterval(t);
  }, [deadline]);
  return secondsLeft;
}
