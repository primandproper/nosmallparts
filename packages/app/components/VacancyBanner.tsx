import { useEffect, useState } from 'react';
import { Pressable, Text, View } from 'react-native';
import type { RoleVacancy } from '@nsp/shared';

interface Props {
  vacancies: RoleVacancy[];
  isAudience: boolean;
  onClaim: (castMemberId: string) => void;
}

/**
 * Shown while a role is vacant mid-performance (§4.5). In non-director mode a
 * live countdown runs against the server's claimDeadline and audience members
 * get a one-click claim button. In director mode (claimDeadline null) the
 * director recasts via the offer flow.
 */
export function VacancyBanner({ vacancies, isAudience, onClaim }: Props) {
  if (vacancies.length === 0) return null;
  return (
    <View>
      {vacancies.map((v) => (
        <VacancyRow key={v.castMemberId} vacancy={v} isAudience={isAudience} onClaim={onClaim} />
      ))}
    </View>
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
  const secondsLeft = useCountdown(vacancy.claimDeadline);
  const expired = vacancy.claimDeadline !== null && secondsLeft <= 0;

  return (
    <View className="border-b border-amber-500 bg-stage-surface px-3 py-2">
      <View className="flex-row flex-wrap items-center justify-between gap-2">
        <View className="flex-1">
          <Text className="text-xs uppercase tracking-widest text-amber-400">Role vacant</Text>
          <Text className="mt-0.5 text-stage-text">
            <Text className="font-semibold">{vacancy.castMemberName}</Text>
            {vacancy.claimDeadline === null
              ? ' needs a new performer. The director will recast the role.'
              : expired
                ? ' went unclaimed.'
                : ` needs a new performer — ${secondsLeft}s left to claim.`}
          </Text>
        </View>
        {isAudience && vacancy.claimDeadline !== null && !expired ? (
          <Pressable
            onPress={() => onClaim(vacancy.castMemberId)}
            accessibilityRole="button"
            accessibilityLabel={`Claim the role ${vacancy.castMemberName}`}
            className="rounded-md bg-stage-gold px-3 py-1.5 active:opacity-80"
          >
            <Text className="text-xs font-semibold text-stage-bg">
              Claim {vacancy.castMemberName}
            </Text>
          </Pressable>
        ) : null}
      </View>
    </View>
  );
}

function useCountdown(deadline: number | null): number {
  const [secondsLeft, setSecondsLeft] = useState(() =>
    deadline === null ? 0 : Math.max(0, Math.ceil((deadline - Date.now()) / 1000))
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
