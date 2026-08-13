import type { Play } from './types.js';

// Sample plays seeded into the DB. Kept deterministic so ids don't drift between
// seeds. The server imports these on startup (or via `pnpm db:seed`).
export const samplePlays: Play[] = [
  {
    id: 'dinner-plans',
    title: 'Dinner Plans',
    description:
      'A short scene about coordinating dinner plans between friends. Perfect for 3 performers.',
    author: 'No Small Parts',
    cast: [
      { id: 'rhonda', name: 'Rhonda' },
      { id: 'charles', name: 'Charles' },
      { id: 'ryan', name: 'Ryan' },
    ],
    lines: [
      { castMemberIds: ['rhonda'], text: 'What time is dinner?', index: 0 },
      { castMemberIds: ['charles'], text: "It's at 7, Rhonda", index: 1 },
      { castMemberIds: ['rhonda'], text: "And it's at the Italian place we like?", index: 2 },
      { castMemberIds: ['ryan'], text: "No, it's at the Greek spot", index: 3 },
    ],
  },
  {
    id: 'morning-coffee',
    title: 'Morning Coffee',
    description: 'A comedic exchange about whose turn it is to make coffee. Great for duos.',
    author: 'No Small Parts',
    cast: [
      { id: 'alex', name: 'Alex' },
      { id: 'jordan', name: 'Jordan' },
    ],
    lines: [
      { castMemberIds: ['alex'], text: 'Did you make coffee this morning?', index: 0 },
      { castMemberIds: ['jordan'], text: 'I thought it was your turn.', index: 1 },
      { castMemberIds: ['alex'], text: 'My turn? I made it yesterday!', index: 2 },
      { castMemberIds: ['jordan'], text: "Fine, I'll make it. But you're buying lunch.", index: 3 },
      { castMemberIds: ['alex'], text: 'Deal.', index: 4 },
    ],
  },
  {
    id: 'the-meeting',
    title: 'The Meeting',
    description:
      'A corporate comedy about status updates and confused interns. Perfect for 4 performers.',
    author: 'No Small Parts',
    cast: [
      { id: 'boss', name: 'The Boss' },
      { id: 'employee1', name: 'Employee 1' },
      { id: 'employee2', name: 'Employee 2' },
      { id: 'intern', name: 'The Intern' },
    ],
    lines: [
      { castMemberIds: ['boss'], text: "Let's get this meeting started. Status updates?", index: 0 },
      { castMemberIds: ['employee1'], text: 'The project is on track, sir.', index: 1 },
      { castMemberIds: ['employee2'], text: "We're slightly behind, but catching up.", index: 2 },
      { castMemberIds: ['boss'], text: 'And the new initiative?', index: 3 },
      { castMemberIds: ['intern'], text: 'Um... what initiative?', index: 4 },
      { castMemberIds: ['boss'], text: "The one from last week's email!", index: 5 },
      { castMemberIds: ['intern'], text: "Oh! Right! It's... progressing?", index: 6 },
      {
        castMemberIds: ['employee1'],
        text: "I'll help them get up to speed after this.",
        index: 7,
      },
      { castMemberIds: ['boss'], text: 'Good. Meeting adjourned.', index: 8 },
    ],
  },
  {
    id: 'chorus-demo',
    title: 'The Chorus',
    description:
      'A demonstration of simultaneous speaking: unison (one line, multiple speakers) and overlapping dialogue (two characters talking over each other).',
    author: 'No Small Parts',
    cast: [
      { id: 'narrator', name: 'Narrator' },
      { id: 'chorus1', name: 'Chorus Member 1' },
      { id: 'chorus2', name: 'Chorus Member 2' },
      { id: 'chorus3', name: 'Chorus Member 3' },
    ],
    lines: [
      { castMemberIds: ['narrator'], text: 'Once upon a time, in a land far away...', index: 0 },
      // Unison: all three speak the same text together.
      {
        castMemberIds: ['chorus1', 'chorus2', 'chorus3'],
        text: 'We gather here today!',
        index: 1,
      },
      { castMemberIds: ['narrator'], text: 'The chorus spoke as one...', index: 2 },
      // Overlap: chorus1+chorus2 say one thing in unison while chorus3 talks over them.
      {
        castMemberIds: ['chorus1', 'chorus2'],
        text: 'To tell a tale of old!',
        index: 3,
      },
      {
        castMemberIds: ['chorus3'],
        text: 'But not too old, mind you!',
        index: 3,
      },
      { castMemberIds: ['narrator'], text: 'And so the story began.', index: 4 },
    ],
  },
  {
    id: 'job-interview',
    title: 'The Interview',
    description:
      'A nervous job candidate faces tough questions. A two-person scene with comedic potential. Demonstrates stage directions.',
    author: 'No Small Parts',
    cast: [
      { id: 'interviewer', name: 'Interviewer' },
      { id: 'candidate', name: 'Candidate' },
    ],
    lines: [
      {
        castMemberIds: ['interviewer'],
        text: 'So, tell me about yourself.',
        index: 0,
        direction: 'looking down at resume',
      },
      {
        castMemberIds: ['candidate'],
        text: "Well, I'm a hard worker, team player, and I love... synergy?",
        index: 1,
        direction: 'nervously',
      },
      {
        castMemberIds: ['interviewer'],
        text: 'Interesting. Where do you see yourself in five years?',
        index: 2,
        direction: 'raising an eyebrow',
      },
      {
        castMemberIds: ['candidate'],
        text: 'Hopefully still employed... I mean, leading a team!',
        index: 3,
        direction: 'catching themselves',
      },
      {
        castMemberIds: ['interviewer'],
        text: "What's your greatest weakness?",
        index: 4,
        direction: 'leaning forward',
      },
      {
        castMemberIds: ['candidate'],
        text: "I work too hard. Also, I'm terrible at answering this question.",
        index: 5,
        direction: 'with a nervous laugh',
      },
      {
        castMemberIds: ['interviewer'],
        text: "At least you're honest. You're hired.",
        index: 6,
        direction: 'smiling warmly',
      },
      {
        castMemberIds: ['candidate'],
        text: 'Wait, really?!',
        index: 7,
        direction: 'shocked, then elated',
      },
    ],
  },
  {
    id: 'romeo-juliet-balcony',
    title: 'Romeo and Juliet - Balcony Scene',
    description:
      "The famous balcony scene from Shakespeare's Romeo and Juliet. A romantic classic for two performers.",
    author: 'William Shakespeare',
    cast: [
      { id: 'romeo', name: 'Romeo' },
      { id: 'juliet', name: 'Juliet' },
    ],
    lines: [
      {
        castMemberIds: ['romeo'],
        text: 'But soft, what light through yonder window breaks?',
        index: 0,
      },
      { castMemberIds: ['romeo'], text: 'It is the east, and Juliet is the sun.', index: 1 },
      { castMemberIds: ['juliet'], text: 'O Romeo, Romeo! Wherefore art thou Romeo?', index: 2 },
      { castMemberIds: ['juliet'], text: 'Deny thy father and refuse thy name.', index: 3 },
      { castMemberIds: ['romeo'], text: 'Shall I hear more, or shall I speak at this?', index: 4 },
      {
        castMemberIds: ['juliet'],
        text: "What's in a name? That which we call a rose by any other name would smell as sweet.",
        index: 5,
      },
      {
        castMemberIds: ['romeo'],
        text: "I take thee at thy word. Call me but love, and I'll be new baptized.",
        index: 6,
      },
    ],
  },
  {
    id: 'waiting-for-godot',
    title: 'Waiting for Godot - Opening',
    description:
      "The opening exchange from Beckett's absurdist masterpiece. Two friends waiting for someone who never arrives.",
    author: 'Samuel Beckett',
    cast: [
      { id: 'vladimir', name: 'Vladimir' },
      { id: 'estragon', name: 'Estragon' },
    ],
    lines: [
      { castMemberIds: ['estragon'], text: 'Nothing to be done.', index: 0 },
      { castMemberIds: ['vladimir'], text: "I'm beginning to come round to that opinion.", index: 1 },
      { castMemberIds: ['estragon'], text: "I've tried everything.", index: 2 },
      { castMemberIds: ['vladimir'], text: 'So have I.', index: 3 },
      { castMemberIds: ['estragon'], text: "Let's go.", index: 4 },
      { castMemberIds: ['vladimir'], text: "We can't.", index: 5 },
      { castMemberIds: ['estragon'], text: 'Why not?', index: 6 },
      { castMemberIds: ['vladimir'], text: "We're waiting for Godot.", index: 7 },
    ],
  },
  {
    id: 'customer-service',
    title: 'Customer Service',
    description: 'A frustrated customer calls a helpline. Comedy gold for two performers.',
    author: 'No Small Parts',
    cast: [
      { id: 'customer', name: 'Customer' },
      { id: 'agent', name: 'Support Agent' },
    ],
    lines: [
      {
        castMemberIds: ['agent'],
        text: 'Thank you for calling support. How may I help you today?',
        index: 0,
      },
      { castMemberIds: ['customer'], text: 'Yes, my internet has been down for THREE DAYS.', index: 1 },
      {
        castMemberIds: ['agent'],
        text: "I'm sorry to hear that. Have you tried turning it off and on again?",
        index: 2,
      },
      { castMemberIds: ['customer'], text: "Of course I have! I've tried everything!", index: 3 },
      {
        castMemberIds: ['agent'],
        text: 'I understand your frustration. Let me check your account...',
        index: 4,
      },
      { castMemberIds: ['agent'], text: 'Ah, I see the problem. Your bill is overdue.', index: 5 },
      { castMemberIds: ['customer'], text: 'What?! I paid it last week!', index: 6 },
      {
        castMemberIds: ['agent'],
        text: "Oh wait, that was someone else's account. You're all good!",
        index: 7,
      },
      { castMemberIds: ['customer'], text: "Then why isn't it working?!", index: 8 },
      { castMemberIds: ['agent'], text: 'Have you tried turning it off and on again?', index: 9 },
    ],
  },
  {
    id: 'mark-antony-funeral',
    title: 'Julius Caesar - Friends, Romans, Countrymen',
    description:
      "Mark Antony's funeral oration over Caesar's body. A single sustained monologue — useful for testing long-line rendering and auto-scroll.",
    author: 'William Shakespeare',
    cast: [
      { id: 'antony', name: 'Mark Antony' },
      { id: 'plebeian', name: 'Plebeian' },
    ],
    lines: [
      {
        castMemberIds: ['antony'],
        text: "Friends, Romans, countrymen, lend me your ears; I come to bury Caesar, not to praise him. The evil that men do lives after them; the good is oft interred with their bones; so let it be with Caesar. The noble Brutus hath told you Caesar was ambitious: if it were so, it was a grievous fault, and grievously hath Caesar answer'd it. Here, under leave of Brutus and the rest— for Brutus is an honourable man; so are they all, all honourable men— come I to speak in Caesar's funeral. He was my friend, faithful and just to me: but Brutus says he was ambitious; and Brutus is an honourable man. He hath brought many captives home to Rome whose ransoms did the general coffers fill: did this in Caesar seem ambitious? When that the poor have cried, Caesar hath wept: ambition should be made of sterner stuff: yet Brutus says he was ambitious; and Brutus is an honourable man. You all did see that on the Lupercal I thrice presented him a kingly crown, which he did thrice refuse: was this ambition? Yet Brutus says he was ambitious; and, sure, he is an honourable man. I speak not to disprove what Brutus spoke, but here I am to speak what I do know. You all did love him once, not without cause: what cause withholds you then, to mourn for him? O judgment! thou art fled to brutish beasts, and men have lost their reason. Bear with me; my heart is in the coffin there with Caesar, and I must pause till it come back to me.",
        index: 0,
        direction: "addressing the crowd over Caesar's bier",
        scrollMode: 'auto',
        autoScrollWpm: 130,
      },
      {
        castMemberIds: ['plebeian'],
        text: 'Methinks there is much reason in his sayings.',
        index: 1,
        direction: 'murmuring',
      },
      {
        castMemberIds: ['antony'],
        text: "But yesterday the word of Caesar might have stood against the world; now lies he there, and none so poor to do him reverence. O masters, if I were disposed to stir your hearts and minds to mutiny and rage, I should do Brutus wrong, and Cassius wrong, who, you all know, are honourable men: I will not do them wrong; I rather choose to wrong the dead, to wrong myself and you, than I will wrong such honourable men. But here's a parchment with the seal of Caesar; I found it in his closet, 'tis his will: let but the commons hear this testament— which, pardon me, I do not mean to read— and they would go and kiss dead Caesar's wounds and dip their napkins in his sacred blood, yea, beg a hair of him for memory, and, dying, mention it within their wills, bequeathing it as a rich legacy unto their issue.",
        index: 2,
        direction: 'producing the will',
        scrollMode: 'auto',
        autoScrollWpm: 130,
      },
      {
        castMemberIds: ['plebeian'],
        text: "We'll hear the will: read it, Mark Antony.",
        index: 3,
        direction: 'shouting',
      },
      {
        castMemberIds: ['antony'],
        text: "Have patience, gentle friends, I must not read it; it is not meet you know how Caesar loved you. You are not wood, you are not stones, but men; and, being men, hearing the will of Caesar, it will inflame you, it will make you mad: 'tis good you know not that you are his heirs; for, if you should, O, what would come of it!",
        index: 4,
        direction: 'feigning reluctance',
      },
    ],
  },
  {
    id: 'hamlet-soliloquy',
    title: 'Hamlet - To Be Or Not To Be',
    description:
      'The famous soliloquy from Hamlet. Demonstrates the auto-scroll feature for long monologues.',
    author: 'William Shakespeare',
    cast: [
      { id: 'hamlet', name: 'Hamlet' },
      { id: 'ophelia', name: 'Ophelia' },
    ],
    lines: [
      {
        castMemberIds: ['hamlet'],
        text: "To be, or not to be, that is the question: Whether 'tis nobler in the mind to suffer the slings and arrows of outrageous fortune, or to take arms against a sea of troubles and by opposing end them. To die—to sleep, no more; and by a sleep to say we end the heart-ache and the thousand natural shocks that flesh is heir to: 'tis a consummation devoutly to be wish'd. To die, to sleep; to sleep, perchance to dream—ay, there's the rub.",
        index: 0,
        direction: 'contemplating, holding skull',
        scrollMode: 'auto',
        autoScrollWpm: 120,
      },
      {
        castMemberIds: ['ophelia'],
        text: 'Good my lord, how does your honour for this many a day?',
        index: 1,
        direction: 'approaching cautiously',
      },
      {
        castMemberIds: ['hamlet'],
        text: 'I humbly thank you; well, well, well.',
        index: 2,
        direction: 'distant, distracted',
      },
    ],
  },
];
