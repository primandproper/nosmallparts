import { Pressable, Text, View } from 'react-native';
import type { RoleOffer } from '@nsp/shared';

interface Props {
  offers: RoleOffer[];
  onRespond: (offerId: string, accept: boolean) => void;
}

export function RoleOfferPrompt({ offers, onRespond }: Props) {
  if (offers.length === 0) return null;
  // Only surface the oldest pending offer; others queue behind.
  const offer = offers[0];
  return (
    <View className="border-b border-stage-gold bg-stage-surface px-3 py-2">
      <Text className="text-xs uppercase tracking-widest text-stage-gold">Role offered</Text>
      <Text className="mt-1 text-stage-text">
        The host has offered you <Text className="font-semibold">{offer.castMemberName}</Text>.
      </Text>
      <Text className="mt-0.5 text-xs text-stage-muted">
        Accepting puts you in the cast — you can enable your camera and mic to perform.
      </Text>
      <View className="mt-2 flex-row gap-2">
        <Pressable
          onPress={() => onRespond(offer.offerId, true)}
          className="rounded-md bg-stage-gold px-3 py-1"
        >
          <Text className="text-stage-bg font-semibold text-xs">Accept</Text>
        </Pressable>
        <Pressable
          onPress={() => onRespond(offer.offerId, false)}
          className="rounded-md border border-stage-border px-3 py-1"
        >
          <Text className="text-stage-muted text-xs">Decline</Text>
        </Pressable>
      </View>
    </View>
  );
}
