import Link from 'next/link';

export const metadata = {
  title: 'Privacy Policy | No Small Parts',
  description: 'Privacy Policy for No Small Parts - A collaborative play performance platform',
};

export default function PrivacyPolicyPage() {
  return (
    <main className="min-h-screen stage-bg relative overflow-hidden py-16 px-6">
      <div className="max-w-3xl mx-auto">
        <Link
          href="/"
          className="inline-flex items-center gap-2 text-[var(--gold)] hover:text-[var(--gold-light)] transition-colors text-sm mb-8"
        >
          ← Back to Home
        </Link>

        <div className="card-elegant rounded-2xl p-8 md:p-12">
          <h1 className="font-[Cormorant_Garamond] text-4xl font-bold text-gold-gradient mb-2">
            Privacy Policy
          </h1>
          <p className="text-[var(--cream)] opacity-60 text-sm mb-8">
            Last updated: February 15, 2025
          </p>

          <div className="prose prose-invert max-w-none text-[var(--cream)] space-y-6 text-sm leading-relaxed">
            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">1. Introduction</h2>
              <p className="opacity-90">
                No Small Parts (&quot;we,&quot; &quot;our,&quot; or &quot;us&quot;) respects your
                privacy and is committed to protecting your personal data. This Privacy Policy
                explains how we collect, use, disclose, and safeguard your information when you use
                our collaborative play performance platform (the &quot;Service&quot;).
              </p>
            </section>

            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">
                2. Information We Collect
              </h2>
              <p className="opacity-90 mb-3">We collect information you provide directly to us:</p>
              <ul className="list-disc pl-6 space-y-2 opacity-90">
                <li>
                  <strong>Account information:</strong> Username, email address, first and last
                  name, and password when you register
                </li>
                <li>
                  <strong>Profile information:</strong> Optional birthday and other details you
                  choose to add
                </li>
                <li>
                  <strong>Content:</strong> Play scripts, performance data, chat messages, and other
                  materials you create or upload
                </li>
                <li>
                  <strong>Communications:</strong> Messages you send through the Service and support
                  requests
                </li>
              </ul>
              <p className="opacity-90 mt-3">
                We also automatically collect certain technical information, including IP address,
                browser type, device information, and usage data, to operate and improve the
                Service.
              </p>
            </section>

            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">
                3. How We Use Your Information
              </h2>
              <p className="opacity-90 mb-3">We use the information we collect to:</p>
              <ul className="list-disc pl-6 space-y-2 opacity-90">
                <li>Provide, maintain, and improve the Service</li>
                <li>Authenticate users and manage accounts</li>
                <li>
                  Enable collaborative features such as stages, plays, and real-time performance
                </li>
                <li>Send transactional emails and service-related notifications</li>
                <li>Respond to support requests and inquiries</li>
                <li>Analyze usage patterns to enhance user experience</li>
                <li>Detect and prevent fraud, abuse, and security incidents</li>
                <li>Comply with legal obligations</li>
              </ul>
            </section>

            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">
                4. Information Sharing
              </h2>
              <p className="opacity-90">
                We do not sell your personal information. We may share your information with: (i)
                service providers who assist us in operating the Service (under confidentiality
                obligations); (ii) other users as necessary for collaborative features (e.g.,
                display name in a stage); (iii) law enforcement or other parties when required by
                law or to protect our rights and safety.
              </p>
            </section>

            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">5. Data Security</h2>
              <p className="opacity-90">
                We implement appropriate technical and organizational measures to protect your
                personal data against unauthorized access, alteration, disclosure, or destruction.
                Passwords are hashed and stored securely. We use industry-standard practices for
                data transmission and storage. However, no method of transmission over the Internet
                is 100% secure, and we cannot guarantee absolute security.
              </p>
            </section>

            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">6. Data Retention</h2>
              <p className="opacity-90">
                We retain your information for as long as your account is active or as needed to
                provide the Service. We may retain certain information after account deletion as
                required by law, for legitimate business purposes, or to resolve disputes.
              </p>
            </section>

            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">7. Your Rights</h2>
              <p className="opacity-90">
                Depending on your location, you may have rights to access, correct, delete, or
                export your personal data, or to object to or restrict certain processing. You can
                update your account information through the Service. To exercise other rights or for
                assistance, please contact us.
              </p>
            </section>

            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">
                8. Cookies and Tracking
              </h2>
              <p className="opacity-90">
                We use cookies and similar technologies to maintain session state, remember
                preferences, and analyze usage. You can configure your browser to refuse cookies,
                but some features of the Service may not function properly without them.
              </p>
            </section>

            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">
                9. Children&apos;s Privacy
              </h2>
              <p className="opacity-90">
                The Service is not intended for users under 18 years of age. We do not knowingly
                collect personal information from children. If you believe we have collected
                information from a minor, please contact us so we can delete it.
              </p>
            </section>

            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">
                10. Changes to This Policy
              </h2>
              <p className="opacity-90">
                We may update this Privacy Policy from time to time. We will notify you of material
                changes by posting the updated policy on the Service and updating the &quot;Last
                updated&quot; date. Your continued use of the Service after such changes constitutes
                acceptance of the revised policy.
              </p>
            </section>

            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">11. Contact Us</h2>
              <p className="opacity-90">
                For questions about this Privacy Policy or our privacy practices, please contact us
                through the support options provided in the Service.
              </p>
            </section>
          </div>

          <div className="mt-12 pt-8 border-t border-[var(--gold)]/20">
            <Link
              href="/auth/register"
              className="text-[var(--gold)] hover:text-[var(--gold-light)] transition-colors text-sm"
            >
              ← Back to Registration
            </Link>
          </div>
        </div>
      </div>
    </main>
  );
}
