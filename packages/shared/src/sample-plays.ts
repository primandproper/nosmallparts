import type { Play } from './types';

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
      { castMemberId: 'rhonda', text: 'What time is dinner?', index: 0 },
      { castMemberId: 'charles', text: "It's at 7, Rhonda", index: 1 },
      { castMemberId: 'rhonda', text: "And it's at the Italian place we like?", index: 2 },
      { castMemberId: 'ryan', text: "No, it's at the Greek spot", index: 3 },
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
      { castMemberId: 'alex', text: 'Did you make coffee this morning?', index: 0 },
      { castMemberId: 'jordan', text: 'I thought it was your turn.', index: 1 },
      { castMemberId: 'alex', text: 'My turn? I made it yesterday!', index: 2 },
      { castMemberId: 'jordan', text: "Fine, I'll make it. But you're buying lunch.", index: 3 },
      { castMemberId: 'alex', text: 'Deal.', index: 4 },
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
      { castMemberId: 'boss', text: "Let's get this meeting started. Status updates?", index: 0 },
      { castMemberId: 'employee1', text: 'The project is on track, sir.', index: 1 },
      { castMemberId: 'employee2', text: "We're slightly behind, but catching up.", index: 2 },
      { castMemberId: 'boss', text: 'And the new initiative?', index: 3 },
      { castMemberId: 'intern', text: 'Um... what initiative?', index: 4 },
      { castMemberId: 'boss', text: "The one from last week's email!", index: 5 },
      { castMemberId: 'intern', text: "Oh! Right! It's... progressing?", index: 6 },
      {
        castMemberId: 'employee1',
        text: "I'll help them get up to speed after this.",
        index: 7,
      },
      { castMemberId: 'boss', text: 'Good. Meeting adjourned.', index: 8 },
    ],
  },
  {
    id: 'chorus-demo',
    title: 'The Chorus',
    description:
      'A demonstration of simultaneous speaking, featuring a narrator and chorus members speaking in unison.',
    author: 'No Small Parts',
    cast: [
      { id: 'narrator', name: 'Narrator' },
      { id: 'chorus1', name: 'Chorus Member 1' },
      { id: 'chorus2', name: 'Chorus Member 2' },
      { id: 'chorus3', name: 'Chorus Member 3' },
    ],
    lines: [
      { castMemberId: 'narrator', text: 'Once upon a time, in a land far away...', index: 0 },
      { castMemberId: 'chorus1', text: 'We gather here today!', index: 1 },
      { castMemberId: 'chorus2', text: 'We gather here today!', index: 1 },
      { castMemberId: 'chorus3', text: 'We gather here today!', index: 1 },
      { castMemberId: 'narrator', text: 'The chorus spoke as one...', index: 2 },
      { castMemberId: 'chorus1', text: 'To tell a tale of old!', index: 3 },
      { castMemberId: 'chorus2', text: 'To tell a tale of old!', index: 3 },
      { castMemberId: 'narrator', text: 'And so the story began.', index: 4 },
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
        castMemberId: 'interviewer',
        text: 'So, tell me about yourself.',
        index: 0,
        direction: 'looking down at resume',
      },
      {
        castMemberId: 'candidate',
        text: "Well, I'm a hard worker, team player, and I love... synergy?",
        index: 1,
        direction: 'nervously',
      },
      {
        castMemberId: 'interviewer',
        text: 'Interesting. Where do you see yourself in five years?',
        index: 2,
        direction: 'raising an eyebrow',
      },
      {
        castMemberId: 'candidate',
        text: 'Hopefully still employed... I mean, leading a team!',
        index: 3,
        direction: 'catching themselves',
      },
      {
        castMemberId: 'interviewer',
        text: "What's your greatest weakness?",
        index: 4,
        direction: 'leaning forward',
      },
      {
        castMemberId: 'candidate',
        text: "I work too hard. Also, I'm terrible at answering this question.",
        index: 5,
        direction: 'with a nervous laugh',
      },
      {
        castMemberId: 'interviewer',
        text: "At least you're honest. You're hired.",
        index: 6,
        direction: 'smiling warmly',
      },
      {
        castMemberId: 'candidate',
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
        castMemberId: 'romeo',
        text: 'But soft, what light through yonder window breaks?',
        index: 0,
      },
      { castMemberId: 'romeo', text: 'It is the east, and Juliet is the sun.', index: 1 },
      { castMemberId: 'juliet', text: 'O Romeo, Romeo! Wherefore art thou Romeo?', index: 2 },
      { castMemberId: 'juliet', text: 'Deny thy father and refuse thy name.', index: 3 },
      { castMemberId: 'romeo', text: 'Shall I hear more, or shall I speak at this?', index: 4 },
      {
        castMemberId: 'juliet',
        text: "What's in a name? That which we call a rose by any other name would smell as sweet.",
        index: 5,
      },
      {
        castMemberId: 'romeo',
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
      { castMemberId: 'estragon', text: 'Nothing to be done.', index: 0 },
      { castMemberId: 'vladimir', text: "I'm beginning to come round to that opinion.", index: 1 },
      { castMemberId: 'estragon', text: "I've tried everything.", index: 2 },
      { castMemberId: 'vladimir', text: 'So have I.', index: 3 },
      { castMemberId: 'estragon', text: "Let's go.", index: 4 },
      { castMemberId: 'vladimir', text: "We can't.", index: 5 },
      { castMemberId: 'estragon', text: 'Why not?', index: 6 },
      { castMemberId: 'vladimir', text: "We're waiting for Godot.", index: 7 },
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
        castMemberId: 'agent',
        text: 'Thank you for calling support. How may I help you today?',
        index: 0,
      },
      { castMemberId: 'customer', text: 'Yes, my internet has been down for THREE DAYS.', index: 1 },
      {
        castMemberId: 'agent',
        text: "I'm sorry to hear that. Have you tried turning it off and on again?",
        index: 2,
      },
      { castMemberId: 'customer', text: "Of course I have! I've tried everything!", index: 3 },
      {
        castMemberId: 'agent',
        text: 'I understand your frustration. Let me check your account...',
        index: 4,
      },
      { castMemberId: 'agent', text: 'Ah, I see the problem. Your bill is overdue.', index: 5 },
      { castMemberId: 'customer', text: 'What?! I paid it last week!', index: 6 },
      {
        castMemberId: 'agent',
        text: "Oh wait, that was someone else's account. You're all good!",
        index: 7,
      },
      { castMemberId: 'customer', text: "Then why isn't it working?!", index: 8 },
      { castMemberId: 'agent', text: 'Have you tried turning it off and on again?', index: 9 },
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
        castMemberId: 'hamlet',
        text: "To be, or not to be, that is the question: Whether 'tis nobler in the mind to suffer the slings and arrows of outrageous fortune, or to take arms against a sea of troubles and by opposing end them. To die—to sleep, no more; and by a sleep to say we end the heart-ache and the thousand natural shocks that flesh is heir to: 'tis a consummation devoutly to be wish'd. To die, to sleep; to sleep, perchance to dream—ay, there's the rub.",
        index: 0,
        direction: 'contemplating, holding skull',
        scrollMode: 'auto',
        autoScrollWpm: 120,
      },
      {
        castMemberId: 'ophelia',
        text: 'Good my lord, how does your honour for this many a day?',
        index: 1,
        direction: 'approaching cautiously',
      },
      {
        castMemberId: 'hamlet',
        text: 'I humbly thank you; well, well, well.',
        index: 2,
        direction: 'distant, distracted',
      },
    ],
  },
];
