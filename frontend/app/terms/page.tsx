import Link from 'next/link';

export const metadata = {
  title: 'Terms of Service | No Small Parts',
  description: 'Terms of Service for No Small Parts - A collaborative play performance platform',
};

export default function TermsOfServicePage() {
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
            Terms of Service
          </h1>
          <p className="text-[var(--cream)] opacity-60 text-sm mb-8">Last updated: February 15, 2025</p>

          <div className="prose prose-invert max-w-none text-[var(--cream)] space-y-6 text-sm leading-relaxed">
            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">1. Acceptance of Terms</h2>
              <p className="opacity-90">
                By accessing or using No Small Parts (&quot;the Service&quot;), you agree to be bound by these Terms
                of Service. If you do not agree to these terms, please do not use the Service. The Service is
                operated by No Small Parts and provides a collaborative platform for play performance,
                rehearsal, and related activities.
              </p>
            </section>

            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">2. Eligibility</h2>
              <p className="opacity-90">
                You must be at least 18 years of age to use this Service. By using the Service, you represent
                and warrant that you meet this age requirement. We reserve the right to terminate accounts of
                users we believe are under 18.
              </p>
            </section>

            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">3. Account Registration</h2>
              <p className="opacity-90">
                To access certain features, you must register for an account. You agree to provide accurate,
                current, and complete information during registration and to update such information as
                necessary. You are responsible for maintaining the confidentiality of your password and for
                all activities that occur under your account.
              </p>
            </section>

            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">4. Acceptable Use</h2>
              <p className="opacity-90 mb-3">
                You agree to use the Service only for lawful purposes and in accordance with these Terms. You
                will not:
              </p>
              <ul className="list-disc pl-6 space-y-2 opacity-90">
                <li>Violate any applicable laws or regulations</li>
                <li>Infringe on the intellectual property rights of others</li>
                <li>Upload content that is offensive, harassing, defamatory, or inappropriate</li>
                <li>Attempt to gain unauthorized access to the Service or other accounts</li>
                <li>Use the Service to distribute malware or engage in harmful activities</li>
                <li>Interfere with or disrupt the Service or its infrastructure</li>
              </ul>
            </section>

            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">5. Content and Intellectual Property</h2>
              <p className="opacity-90">
                You retain ownership of content you create and upload to the Service. By uploading content,
                you grant us a limited license to host, display, and process your content as necessary to
                provide the Service. The Service itself, including its design, features, and branding, is
                owned by No Small Parts and protected by intellectual property laws. Play scripts and
                theatrical works may be subject to additional copyright; you are responsible for ensuring
                you have appropriate rights to use any materials you upload.
              </p>
            </section>

            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">6. Disclaimer of Warranties</h2>
              <p className="opacity-90">
                The Service is provided &quot;as is&quot; and &quot;as available&quot; without warranties of any kind, express
                or implied. We do not warrant that the Service will be uninterrupted, error-free, or secure.
                Your use of the Service is at your sole risk.
              </p>
            </section>

            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">7. Limitation of Liability</h2>
              <p className="opacity-90">
                To the maximum extent permitted by law, No Small Parts shall not be liable for any indirect,
                incidental, special, consequential, or punitive damages arising from your use of the Service.
                Our total liability shall not exceed the amount you paid (if any) for the Service in the
                twelve months preceding the claim.
              </p>
            </section>

            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">8. Termination</h2>
              <p className="opacity-90">
                We may suspend or terminate your account at any time for violations of these Terms or for any
                other reason. You may terminate your account at any time. Upon termination, your right to use
                the Service ceases immediately.
              </p>
            </section>

            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">9. Changes to Terms</h2>
              <p className="opacity-90">
                We may modify these Terms at any time. We will notify you of material changes by posting the
                updated Terms on the Service and updating the &quot;Last updated&quot; date. Your continued use of the
                Service after such changes constitutes acceptance of the revised Terms.
              </p>
            </section>

            <section>
              <h2 className="text-[var(--gold)] text-lg font-semibold mb-3">10. Contact</h2>
              <p className="opacity-90">
                For questions about these Terms of Service, please contact us through the support options
                provided in the Service.
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
