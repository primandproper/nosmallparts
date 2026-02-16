import { AdBanner } from './AdBanner';

/**
 * Server Component that reads AdSense config from env and renders AdBanner.
 * Use this instead of AdBanner when you want runtime config (e.g. from Terraform/Cloud Run).
 *
 * For GKE/Skaffold: ADSENSE_PUBLISHER_ID and ADSENSE_SLOT_ID must be in the deployment
 * env (see frontend/deploy kustomize patches). They are NOT passed by default.
 */
export function AdBannerServer({
  className = '',
  format = 'auto' as const,
  responsive = true,
}: {
  className?: string;
  format?: 'auto' | 'horizontal' | 'vertical' | 'rectangle';
  responsive?: boolean;
}) {
  const client =
    process.env.ADSENSE_PUBLISHER_ID || process.env.NEXT_PUBLIC_ADSENSE_PUBLISHER_ID || '';
  const slot = process.env.ADSENSE_SLOT_ID || process.env.NEXT_PUBLIC_ADSENSE_SLOT_ID || '';

  if (!client || !slot) return null;

  return (
    <AdBanner
      client={client}
      slot={slot}
      format={format}
      responsive={responsive}
      className={className}
    />
  );
}
