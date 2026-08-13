import { relations, sql } from 'drizzle-orm';
import {
  check,
  integer,
  primaryKey,
  sqliteTable,
  text,
  uniqueIndex,
} from 'drizzle-orm/sqlite-core';

// ============================================================================
// Better Auth tables
// These match Better Auth's expected schema. Managed by Better Auth at runtime.
// ============================================================================

export const user = sqliteTable('user', {
  id: text('id').primaryKey(),
  name: text('name').notNull(),
  email: text('email').notNull().unique(),
  emailVerified: integer('emailVerified', { mode: 'boolean' }).notNull().default(false),
  image: text('image'),
  createdAt: integer('createdAt', { mode: 'timestamp_ms' }).notNull(),
  updatedAt: integer('updatedAt', { mode: 'timestamp_ms' }).notNull(),
});

export const session = sqliteTable('session', {
  id: text('id').primaryKey(),
  userId: text('userId')
    .notNull()
    .references(() => user.id, { onDelete: 'cascade' }),
  token: text('token').notNull().unique(),
  expiresAt: integer('expiresAt', { mode: 'timestamp_ms' }).notNull(),
  ipAddress: text('ipAddress'),
  userAgent: text('userAgent'),
  createdAt: integer('createdAt', { mode: 'timestamp_ms' }).notNull(),
  updatedAt: integer('updatedAt', { mode: 'timestamp_ms' }).notNull(),
});

export const account = sqliteTable('account', {
  id: text('id').primaryKey(),
  userId: text('userId')
    .notNull()
    .references(() => user.id, { onDelete: 'cascade' }),
  accountId: text('accountId').notNull(),
  providerId: text('providerId').notNull(),
  accessToken: text('accessToken'),
  refreshToken: text('refreshToken'),
  idToken: text('idToken'),
  accessTokenExpiresAt: integer('accessTokenExpiresAt', { mode: 'timestamp_ms' }),
  refreshTokenExpiresAt: integer('refreshTokenExpiresAt', { mode: 'timestamp_ms' }),
  scope: text('scope'),
  password: text('password'),
  createdAt: integer('createdAt', { mode: 'timestamp_ms' }).notNull(),
  updatedAt: integer('updatedAt', { mode: 'timestamp_ms' }).notNull(),
});

export const verification = sqliteTable('verification', {
  id: text('id').primaryKey(),
  identifier: text('identifier').notNull(),
  value: text('value').notNull(),
  expiresAt: integer('expiresAt', { mode: 'timestamp_ms' }).notNull(),
  createdAt: integer('createdAt', { mode: 'timestamp_ms' }),
  updatedAt: integer('updatedAt', { mode: 'timestamp_ms' }),
});

// ============================================================================
// Application tables
// ============================================================================

export const playScripts = sqliteTable('play_scripts', {
  id: text('id').primaryKey(),
  title: text('title').notNull(),
  description: text('description').notNull().default(''),
  author: text('author'),
  isPublic: integer('is_public', { mode: 'boolean' }).notNull().default(false),
  enabled: integer('enabled', { mode: 'boolean' }).notNull().default(true),
  createdByUser: text('created_by_user')
    .notNull()
    .references(() => user.id, { onDelete: 'cascade' }),
  createdAt: integer('created_at', { mode: 'timestamp_ms' })
    .notNull()
    .$defaultFn(() => new Date()),
  updatedAt: integer('updated_at', { mode: 'timestamp_ms' })
    .notNull()
    .$defaultFn(() => new Date()),
});

export const playScriptCastMembers = sqliteTable('play_script_cast_members', {
  id: text('id').primaryKey(),
  playScriptId: text('play_script_id')
    .notNull()
    .references(() => playScripts.id, { onDelete: 'cascade' }),
  name: text('name').notNull(),
  description: text('description'),
  sortOrder: integer('sort_order').notNull().default(0),
  createdAt: integer('created_at', { mode: 'timestamp_ms' })
    .notNull()
    .$defaultFn(() => new Date()),
});

export const playScriptLines = sqliteTable('play_script_lines', {
  id: text('id').primaryKey(),
  playScriptId: text('play_script_id')
    .notNull()
    .references(() => playScripts.id, { onDelete: 'cascade' }),
  lineText: text('line_text').notNull(),
  lineIndex: integer('line_index').notNull(),
  direction: text('direction'),
  scrollMode: text('scroll_mode', { enum: ['auto', 'manual'] }).default('auto'),
  autoScrollWpm: integer('auto_scroll_wpm'),
  createdAt: integer('created_at', { mode: 'timestamp_ms' })
    .notNull()
    .$defaultFn(() => new Date()),
});

// Join: which cast members speak a given line. >1 row per line = unison.
export const playScriptLineSpeakers = sqliteTable(
  'play_script_line_speakers',
  {
    lineId: text('line_id')
      .notNull()
      .references(() => playScriptLines.id, { onDelete: 'cascade' }),
    castMemberId: text('cast_member_id')
      .notNull()
      .references(() => playScriptCastMembers.id, { onDelete: 'cascade' }),
  },
  (t) => [primaryKey({ columns: [t.lineId, t.castMemberId] })]
);

export const scriptTags = sqliteTable('script_tags', {
  id: text('id').primaryKey(),
  name: text('name').notNull().unique(),
  description: text('description'),
  createdAt: integer('created_at', { mode: 'timestamp_ms' })
    .notNull()
    .$defaultFn(() => new Date()),
});

export const playScriptTags = sqliteTable(
  'play_script_tags',
  {
    playScriptId: text('play_script_id')
      .notNull()
      .references(() => playScripts.id, { onDelete: 'cascade' }),
    scriptTagId: text('script_tag_id')
      .notNull()
      .references(() => scriptTags.id, { onDelete: 'cascade' }),
  },
  (t) => [primaryKey({ columns: [t.playScriptId, t.scriptTagId] })]
);

export const playScriptRatings = sqliteTable(
  'play_script_ratings',
  {
    id: text('id').primaryKey(),
    playScriptId: text('play_script_id')
      .notNull()
      .references(() => playScripts.id, { onDelete: 'cascade' }),
    userId: text('user_id')
      .notNull()
      .references(() => user.id, { onDelete: 'cascade' }),
    rating: integer('rating').notNull(),
    createdAt: integer('created_at', { mode: 'timestamp_ms' })
      .notNull()
      .$defaultFn(() => new Date()),
    updatedAt: integer('updated_at', { mode: 'timestamp_ms' })
      .notNull()
      .$defaultFn(() => new Date()),
  },
  (t) => [
    uniqueIndex('play_script_ratings_unique').on(t.playScriptId, t.userId),
    check('rating_range', sql`${t.rating} BETWEEN 1 AND 5`),
  ]
);

export const stages = sqliteTable('stages', {
  id: text('id').primaryKey(),
  code: text('code').notNull().unique(),
  playScriptId: text('play_script_id').references(() => playScripts.id, {
    onDelete: 'set null',
  }),
  currentLineIndex: integer('current_line_index').notNull().default(0),
  phase: text('phase', { enum: ['lobby', 'countdown', 'performing', 'discussion'] })
    .notNull()
    .default('lobby'),
  hostUserId: text('host_user_id').references(() => user.id, { onDelete: 'set null' }),
  hostGuestId: text('host_guest_id'),
  pausedByHost: integer('paused_by_host', { mode: 'boolean' }).notNull().default(false),
  directorMode: integer('director_mode', { mode: 'boolean' }).notNull().default(false),
  createdAt: integer('created_at', { mode: 'timestamp_ms' })
    .notNull()
    .$defaultFn(() => new Date()),
  updatedAt: integer('updated_at', { mode: 'timestamp_ms' })
    .notNull()
    .$defaultFn(() => new Date()),
});

export const stageParticipants = sqliteTable('stage_participants', {
  id: text('id').primaryKey(),
  stageId: text('stage_id')
    .notNull()
    .references(() => stages.id, { onDelete: 'cascade' }),
  userId: text('user_id').references(() => user.id, { onDelete: 'set null' }),
  guestId: text('guest_id'),
  displayName: text('display_name').notNull(),
  isReady: integer('is_ready', { mode: 'boolean' }).notNull().default(false),
  isAudience: integer('is_audience', { mode: 'boolean' }).notNull().default(false),
  forceMutedByHost: integer('force_muted_by_host', { mode: 'boolean' })
    .notNull()
    .default(false),
  createdAt: integer('created_at', { mode: 'timestamp_ms' })
    .notNull()
    .$defaultFn(() => new Date()),
  leftAt: integer('left_at', { mode: 'timestamp_ms' }),
});

export const stageBans = sqliteTable('stage_bans', {
  id: text('id').primaryKey(),
  stageId: text('stage_id')
    .notNull()
    .references(() => stages.id, { onDelete: 'cascade' }),
  userId: text('user_id').references(() => user.id, { onDelete: 'set null' }),
  guestId: text('guest_id'),
  bannedAt: integer('banned_at', { mode: 'timestamp_ms' })
    .notNull()
    .$defaultFn(() => new Date()),
});

export const stageIpBans = sqliteTable('stage_ip_bans', {
  id: text('id').primaryKey(),
  stageId: text('stage_id')
    .notNull()
    .references(() => stages.id, { onDelete: 'cascade' }),
  ip: text('ip').notNull(),
  bannedAt: integer('banned_at', { mode: 'timestamp_ms' })
    .notNull()
    .$defaultFn(() => new Date()),
});

export const stageParticipantRoles = sqliteTable(
  'stage_participant_roles',
  {
    stageParticipantId: text('stage_participant_id')
      .notNull()
      .references(() => stageParticipants.id, { onDelete: 'cascade' }),
    castMemberId: text('cast_member_id').notNull(),
  },
  (t) => [primaryKey({ columns: [t.stageParticipantId, t.castMemberId] })]
);

// ============================================================================
// Relations
// ============================================================================

export const userRelations = relations(user, ({ many }) => ({
  playScripts: many(playScripts),
  ratings: many(playScriptRatings),
  stageParticipants: many(stageParticipants),
}));

export const playScriptsRelations = relations(playScripts, ({ one, many }) => ({
  createdBy: one(user, {
    fields: [playScripts.createdByUser],
    references: [user.id],
  }),
  castMembers: many(playScriptCastMembers),
  lines: many(playScriptLines),
  tags: many(playScriptTags),
  ratings: many(playScriptRatings),
}));

export const playScriptCastMembersRelations = relations(playScriptCastMembers, ({ one, many }) => ({
  playScript: one(playScripts, {
    fields: [playScriptCastMembers.playScriptId],
    references: [playScripts.id],
  }),
  lineSpeakers: many(playScriptLineSpeakers),
}));

export const playScriptLinesRelations = relations(playScriptLines, ({ one, many }) => ({
  playScript: one(playScripts, {
    fields: [playScriptLines.playScriptId],
    references: [playScripts.id],
  }),
  speakers: many(playScriptLineSpeakers),
}));

export const playScriptLineSpeakersRelations = relations(playScriptLineSpeakers, ({ one }) => ({
  line: one(playScriptLines, {
    fields: [playScriptLineSpeakers.lineId],
    references: [playScriptLines.id],
  }),
  castMember: one(playScriptCastMembers, {
    fields: [playScriptLineSpeakers.castMemberId],
    references: [playScriptCastMembers.id],
  }),
}));

export const scriptTagsRelations = relations(scriptTags, ({ many }) => ({
  playScriptTags: many(playScriptTags),
}));

export const playScriptTagsRelations = relations(playScriptTags, ({ one }) => ({
  playScript: one(playScripts, {
    fields: [playScriptTags.playScriptId],
    references: [playScripts.id],
  }),
  scriptTag: one(scriptTags, {
    fields: [playScriptTags.scriptTagId],
    references: [scriptTags.id],
  }),
}));

export const playScriptRatingsRelations = relations(playScriptRatings, ({ one }) => ({
  playScript: one(playScripts, {
    fields: [playScriptRatings.playScriptId],
    references: [playScripts.id],
  }),
  user: one(user, {
    fields: [playScriptRatings.userId],
    references: [user.id],
  }),
}));

export const stagesRelations = relations(stages, ({ one, many }) => ({
  playScript: one(playScripts, {
    fields: [stages.playScriptId],
    references: [playScripts.id],
  }),
  participants: many(stageParticipants),
}));

export const stageParticipantsRelations = relations(stageParticipants, ({ one, many }) => ({
  stage: one(stages, {
    fields: [stageParticipants.stageId],
    references: [stages.id],
  }),
  user: one(user, {
    fields: [stageParticipants.userId],
    references: [user.id],
  }),
  roles: many(stageParticipantRoles),
}));

export const stageParticipantRolesRelations = relations(stageParticipantRoles, ({ one }) => ({
  participant: one(stageParticipants, {
    fields: [stageParticipantRoles.stageParticipantId],
    references: [stageParticipants.id],
  }),
}));
