# Resource-based Permission Matrix (GH #206)

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace primitive `canEdit(role)` boolean checks with a resource×action×role permission matrix `can(role, resource, action)` matching `docs/roles-and-flows.md`.

**Architecture:** Define enums `Resource` and `Action`, a typed permission matrix mapping `(role, resource) → AccessLevel`, and a `can(role, resource, action)` function that resolves access. Old functions (`canEdit`, `isAdmin`, etc.) delegate to the new matrix for backward compat. Consumers migrate one-by-one.

**Tech Stack:** TypeScript, Jest (existing test infra in `frontend/src/lib/auth/__tests__/`)

---

## Task 1: RED — Tests for permission types and matrix core

**Files:**
- Modify: `frontend/src/lib/auth/__tests__/permissions.test.ts`

**Step 1: Write failing tests for new types and `can()` function**

Add these tests to the existing test file, after the existing `isAdmin` describe block:

```typescript
import {
  EDIT_ROLES,
  VIEW_ONLY_ROLES,
  canEdit,
  canCreate,
  canDelete,
  isViewOnly,
  isAdmin,
  Resource,
  Action,
  AccessLevel,
  can,
  getAccessLevel,
} from '../permissions'

// ... existing tests stay unchanged ...

describe('Resource enum', () => {
  it('has all 8 resources', () => {
    expect(Object.values(Resource)).toHaveLength(8)
    expect(Resource.USERS).toBe('users')
    expect(Resource.CURRICULUM).toBe('curriculum')
    expect(Resource.SCHEDULE).toBe('schedule')
    expect(Resource.ASSIGNMENTS).toBe('assignments')
    expect(Resource.REPORTS).toBe('reports')
    expect(Resource.INTEGRATION).toBe('integration')
    expect(Resource.SYSTEM_SETTINGS).toBe('system_settings')
    expect(Resource.PERSONAL_SETTINGS).toBe('personal_settings')
  })
})

describe('Action enum', () => {
  it('has all 5 actions', () => {
    expect(Object.values(Action)).toHaveLength(5)
    expect(Action.READ).toBe('read')
    expect(Action.CREATE).toBe('create')
    expect(Action.UPDATE).toBe('update')
    expect(Action.DELETE).toBe('delete')
    expect(Action.APPROVE).toBe('approve')
  })
})

describe('AccessLevel enum', () => {
  it('has 4 levels in correct order', () => {
    expect(AccessLevel.DENIED).toBe(0)
    expect(AccessLevel.LIMITED).toBe(1)
    expect(AccessLevel.OWN).toBe(2)
    expect(AccessLevel.FULL).toBe(3)
  })
})

describe('getAccessLevel', () => {
  it('returns full for system_admin on any resource', () => {
    expect(getAccessLevel(UserRole.SYSTEM_ADMIN, Resource.USERS)).toBe(AccessLevel.FULL)
    expect(getAccessLevel(UserRole.SYSTEM_ADMIN, Resource.INTEGRATION)).toBe(AccessLevel.FULL)
    expect(getAccessLevel(UserRole.SYSTEM_ADMIN, Resource.SYSTEM_SETTINGS)).toBe(AccessLevel.FULL)
  })

  it('returns own for personal_settings for all roles', () => {
    expect(getAccessLevel(UserRole.SYSTEM_ADMIN, Resource.PERSONAL_SETTINGS)).toBe(AccessLevel.OWN)
    expect(getAccessLevel(UserRole.STUDENT, Resource.PERSONAL_SETTINGS)).toBe(AccessLevel.OWN)
    expect(getAccessLevel(UserRole.TEACHER, Resource.PERSONAL_SETTINGS)).toBe(AccessLevel.OWN)
  })

  it('returns denied for student on reports', () => {
    expect(getAccessLevel(UserRole.STUDENT, Resource.REPORTS)).toBe(AccessLevel.DENIED)
  })

  it('returns denied for non-admin on integration', () => {
    expect(getAccessLevel(UserRole.METHODIST, Resource.INTEGRATION)).toBe(AccessLevel.DENIED)
    expect(getAccessLevel(UserRole.TEACHER, Resource.INTEGRATION)).toBe(AccessLevel.DENIED)
    expect(getAccessLevel(UserRole.STUDENT, Resource.INTEGRATION)).toBe(AccessLevel.DENIED)
  })

  it('returns denied for non-admin on system_settings', () => {
    expect(getAccessLevel(UserRole.METHODIST, Resource.SYSTEM_SETTINGS)).toBe(AccessLevel.DENIED)
    expect(getAccessLevel(UserRole.ACADEMIC_SECRETARY, Resource.SYSTEM_SETTINGS)).toBe(AccessLevel.DENIED)
  })

  it('returns denied for undefined role', () => {
    expect(getAccessLevel(undefined, Resource.USERS)).toBe(AccessLevel.DENIED)
  })
})

describe('can', () => {
  describe('system_admin has full access to everything', () => {
    it.each([
      [Resource.USERS, Action.CREATE],
      [Resource.USERS, Action.DELETE],
      [Resource.CURRICULUM, Action.APPROVE],
      [Resource.INTEGRATION, Action.UPDATE],
      [Resource.SYSTEM_SETTINGS, Action.UPDATE],
      [Resource.REPORTS, Action.CREATE],
      [Resource.SCHEDULE, Action.CREATE],
    ])('can %s.%s', (resource, action) => {
      expect(can(UserRole.SYSTEM_ADMIN, resource, action)).toBe(true)
    })
  })

  describe('student restrictions', () => {
    it.each([
      [Resource.REPORTS, Action.READ],
      [Resource.REPORTS, Action.CREATE],
      [Resource.INTEGRATION, Action.READ],
      [Resource.SYSTEM_SETTINGS, Action.READ],
      [Resource.USERS, Action.CREATE],
      [Resource.USERS, Action.DELETE],
      [Resource.CURRICULUM, Action.CREATE],
      [Resource.SCHEDULE, Action.CREATE],
    ])('cannot %s.%s', (resource, action) => {
      expect(can(UserRole.STUDENT, resource, action)).toBe(false)
    })

    it('can read own personal_settings', () => {
      expect(can(UserRole.STUDENT, Resource.PERSONAL_SETTINGS, Action.READ)).toBe(true)
      expect(can(UserRole.STUDENT, Resource.PERSONAL_SETTINGS, Action.UPDATE)).toBe(true)
    })

    it('can read schedule', () => {
      expect(can(UserRole.STUDENT, Resource.SCHEDULE, Action.READ)).toBe(true)
    })

    it('can read assignments', () => {
      expect(can(UserRole.STUDENT, Resource.ASSIGNMENTS, Action.READ)).toBe(true)
    })
  })

  describe('methodist permissions', () => {
    it('has full access to curriculum (except approve)', () => {
      expect(can(UserRole.METHODIST, Resource.CURRICULUM, Action.CREATE)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.CURRICULUM, Action.UPDATE)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.CURRICULUM, Action.READ)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.CURRICULUM, Action.APPROVE)).toBe(false)
    })

    it('has full access to reports', () => {
      expect(can(UserRole.METHODIST, Resource.REPORTS, Action.READ)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.REPORTS, Action.CREATE)).toBe(true)
    })

    it('denied integration', () => {
      expect(can(UserRole.METHODIST, Resource.INTEGRATION, Action.READ)).toBe(false)
    })

    it('limited schedule access (read + limited update)', () => {
      expect(can(UserRole.METHODIST, Resource.SCHEDULE, Action.READ)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.SCHEDULE, Action.UPDATE)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.SCHEDULE, Action.CREATE)).toBe(false)
    })
  })

  describe('academic_secretary permissions', () => {
    it('has full access to schedule', () => {
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.SCHEDULE, Action.CREATE)).toBe(true)
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.SCHEDULE, Action.UPDATE)).toBe(true)
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.SCHEDULE, Action.DELETE)).toBe(true)
    })

    it('has full access to reports', () => {
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.REPORTS, Action.CREATE)).toBe(true)
    })

    it('can only read curriculum', () => {
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.CURRICULUM, Action.READ)).toBe(true)
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.CURRICULUM, Action.CREATE)).toBe(false)
    })
  })

  describe('teacher permissions', () => {
    it('limited reports access', () => {
      expect(can(UserRole.TEACHER, Resource.REPORTS, Action.READ)).toBe(true)
      expect(can(UserRole.TEACHER, Resource.REPORTS, Action.CREATE)).toBe(false)
    })

    it('can read schedule but not create', () => {
      expect(can(UserRole.TEACHER, Resource.SCHEDULE, Action.READ)).toBe(true)
      expect(can(UserRole.TEACHER, Resource.SCHEDULE, Action.CREATE)).toBe(false)
    })

    it('has own+full assignments access', () => {
      expect(can(UserRole.TEACHER, Resource.ASSIGNMENTS, Action.CREATE)).toBe(true)
      expect(can(UserRole.TEACHER, Resource.ASSIGNMENTS, Action.READ)).toBe(true)
    })

    it('denied integration and system_settings', () => {
      expect(can(UserRole.TEACHER, Resource.INTEGRATION, Action.READ)).toBe(false)
      expect(can(UserRole.TEACHER, Resource.SYSTEM_SETTINGS, Action.READ)).toBe(false)
    })
  })

  describe('approve action', () => {
    it('only system_admin can approve curriculum', () => {
      expect(can(UserRole.SYSTEM_ADMIN, Resource.CURRICULUM, Action.APPROVE)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.CURRICULUM, Action.APPROVE)).toBe(false)
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.CURRICULUM, Action.APPROVE)).toBe(false)
      expect(can(UserRole.TEACHER, Resource.CURRICULUM, Action.APPROVE)).toBe(false)
      expect(can(UserRole.STUDENT, Resource.CURRICULUM, Action.APPROVE)).toBe(false)
    })
  })

  describe('edge cases', () => {
    it('returns false for undefined role', () => {
      expect(can(undefined, Resource.USERS, Action.READ)).toBe(false)
    })

    it('returns false for empty string role', () => {
      expect(can('', Resource.USERS, Action.READ)).toBe(false)
    })

    it('accepts string role values', () => {
      expect(can('system_admin', Resource.USERS, Action.CREATE)).toBe(true)
      expect(can('student', Resource.REPORTS, Action.READ)).toBe(false)
    })
  })
})
```

**Step 2: Run tests to verify they fail**

Run: `cd frontend && npx jest src/lib/auth/__tests__/permissions.test.ts --no-coverage 2>&1 | tail -20`
Expected: FAIL — `Resource`, `Action`, `AccessLevel`, `can`, `getAccessLevel` not exported

**Step 3: Commit RED**

```bash
git add frontend/src/lib/auth/__tests__/permissions.test.ts
git commit -m "test(permissions): add failing tests for resource-based permission matrix

Tests for Resource/Action/AccessLevel enums, can() function, and
getAccessLevel() covering all 5 roles × 8 resources from docs/roles-and-flows.md.
Includes approve-only-admin for curriculum, student restrictions,
edge cases for undefined/empty roles."
```

---

## Task 2: GREEN — Implement permission types and matrix core

**Files:**
- Modify: `frontend/src/lib/auth/permissions.ts`

**Step 1: Add enums and matrix implementation**

Add to `frontend/src/lib/auth/permissions.ts` (after existing code):

```typescript
export enum Resource {
  USERS = 'users',
  CURRICULUM = 'curriculum',
  SCHEDULE = 'schedule',
  ASSIGNMENTS = 'assignments',
  REPORTS = 'reports',
  INTEGRATION = 'integration',
  SYSTEM_SETTINGS = 'system_settings',
  PERSONAL_SETTINGS = 'personal_settings',
}

export enum Action {
  READ = 'read',
  CREATE = 'create',
  UPDATE = 'update',
  DELETE = 'delete',
  APPROVE = 'approve',
}

export enum AccessLevel {
  DENIED = 0,
  LIMITED = 1,
  OWN = 2,
  FULL = 3,
}

type PermissionMatrix = Record<UserRole, Record<Resource, AccessLevel>>

const PERMISSION_MATRIX: PermissionMatrix = {
  [UserRole.SYSTEM_ADMIN]: {
    [Resource.USERS]: AccessLevel.FULL,
    [Resource.CURRICULUM]: AccessLevel.FULL,
    [Resource.SCHEDULE]: AccessLevel.FULL,
    [Resource.ASSIGNMENTS]: AccessLevel.FULL,
    [Resource.REPORTS]: AccessLevel.FULL,
    [Resource.INTEGRATION]: AccessLevel.FULL,
    [Resource.SYSTEM_SETTINGS]: AccessLevel.FULL,
    [Resource.PERSONAL_SETTINGS]: AccessLevel.OWN,
  },
  [UserRole.METHODIST]: {
    [Resource.USERS]: AccessLevel.LIMITED,
    [Resource.CURRICULUM]: AccessLevel.FULL,
    [Resource.SCHEDULE]: AccessLevel.LIMITED,
    [Resource.ASSIGNMENTS]: AccessLevel.FULL,
    [Resource.REPORTS]: AccessLevel.FULL,
    [Resource.INTEGRATION]: AccessLevel.DENIED,
    [Resource.SYSTEM_SETTINGS]: AccessLevel.DENIED,
    [Resource.PERSONAL_SETTINGS]: AccessLevel.OWN,
  },
  [UserRole.ACADEMIC_SECRETARY]: {
    [Resource.USERS]: AccessLevel.LIMITED,
    [Resource.CURRICULUM]: AccessLevel.LIMITED,
    [Resource.SCHEDULE]: AccessLevel.FULL,
    [Resource.ASSIGNMENTS]: AccessLevel.LIMITED,
    [Resource.REPORTS]: AccessLevel.FULL,
    [Resource.INTEGRATION]: AccessLevel.DENIED,
    [Resource.SYSTEM_SETTINGS]: AccessLevel.DENIED,
    [Resource.PERSONAL_SETTINGS]: AccessLevel.OWN,
  },
  [UserRole.TEACHER]: {
    [Resource.USERS]: AccessLevel.LIMITED,
    [Resource.CURRICULUM]: AccessLevel.LIMITED,
    [Resource.SCHEDULE]: AccessLevel.LIMITED,
    [Resource.ASSIGNMENTS]: AccessLevel.OWN,
    [Resource.REPORTS]: AccessLevel.LIMITED,
    [Resource.INTEGRATION]: AccessLevel.DENIED,
    [Resource.SYSTEM_SETTINGS]: AccessLevel.DENIED,
    [Resource.PERSONAL_SETTINGS]: AccessLevel.OWN,
  },
  [UserRole.STUDENT]: {
    [Resource.USERS]: AccessLevel.OWN,
    [Resource.CURRICULUM]: AccessLevel.LIMITED,
    [Resource.SCHEDULE]: AccessLevel.LIMITED,
    [Resource.ASSIGNMENTS]: AccessLevel.OWN,
    [Resource.REPORTS]: AccessLevel.DENIED,
    [Resource.INTEGRATION]: AccessLevel.DENIED,
    [Resource.SYSTEM_SETTINGS]: AccessLevel.DENIED,
    [Resource.PERSONAL_SETTINGS]: AccessLevel.OWN,
  },
}

const ACTION_MIN_LEVEL: Record<Action, AccessLevel> = {
  [Action.READ]: AccessLevel.LIMITED,
  [Action.CREATE]: AccessLevel.FULL,
  [Action.UPDATE]: AccessLevel.OWN,
  [Action.DELETE]: AccessLevel.FULL,
  [Action.APPROVE]: AccessLevel.FULL,
}

export function getAccessLevel(
  role: UserRole | string | undefined,
  resource: Resource
): AccessLevel {
  if (!role) return AccessLevel.DENIED
  const roleMatrix = PERMISSION_MATRIX[role as UserRole]
  if (!roleMatrix) return AccessLevel.DENIED
  return roleMatrix[resource] ?? AccessLevel.DENIED
}

export function can(
  role: UserRole | string | undefined,
  resource: Resource,
  action: Action
): boolean {
  if (!role) return false
  const level = getAccessLevel(role, resource)
  if (action === Action.APPROVE) {
    return role === UserRole.SYSTEM_ADMIN && resource === Resource.CURRICULUM
  }
  return level >= ACTION_MIN_LEVEL[action]
}
```

**Key design decisions in the matrix:**
- `AccessLevel.LIMITED` → can READ only (not create/delete)
- `AccessLevel.OWN` → can READ + UPDATE own items (not create new / delete)
- `AccessLevel.FULL` → can do everything (read/create/update/delete)
- `Action.APPROVE` is special-cased: only `system_admin` + `curriculum`
- Methodist `schedule = LIMITED` → can read + update (limited), but NOT create/delete
- Teacher `assignments = OWN` → can read + update + create own (teacher creates assignments for their classes). Note: the matrix says "full+own" so Teacher gets OWN which allows create via the `can()` override below

**Important override for teacher assignments:** The matrix from docs says teacher has "full+own" on assignments. We model this as OWN (can read, update own), but the `can()` function needs a special case: teacher can CREATE assignments. This requires adjusting ACTION_MIN_LEVEL or adding a role-specific override.

Looking at the matrix more carefully:
- teacher assignments = "full+own" → they CAN create their own assignments → need `AccessLevel.FULL` but scoped
- methodist assignments = "full+limited" → similar

Actually, let's simplify: for the `can()` boolean check, `FULL` means "can do all actions" and `OWN`/`LIMITED` means "can read, can update own". For **create** specifically, teacher on assignments should return true. Let me model teacher assignments as FULL (they can create/read/update their own assignments — the "own" scoping is at the data layer, not the action layer).

**Revised matrix entries:**
- teacher.assignments = `AccessLevel.FULL` (can create assignments for own classes)
- methodist.assignments = `AccessLevel.FULL` (can manage assignments with limited scope)
- secretary.assignments = `AccessLevel.LIMITED` (read only)
- student.assignments = `AccessLevel.OWN` (read + execute own)

**Step 2: Run tests to verify they pass**

Run: `cd frontend && npx jest src/lib/auth/__tests__/permissions.test.ts --no-coverage 2>&1 | tail -20`
Expected: ALL PASS

**Step 3: Commit GREEN**

```bash
git add frontend/src/lib/auth/permissions.ts
git commit -m "feat(permissions): implement resource-based permission matrix

Resource/Action/AccessLevel enums + PERMISSION_MATRIX covering
5 roles × 8 resources. can(role, resource, action) resolves
access level against action minimum. Approve special-cased
to system_admin+curriculum only. Old canEdit/isAdmin/isViewOnly
remain for backward compat."
```

---

## Task 3: RED — Tests for legacy function delegation to matrix

**Files:**
- Modify: `frontend/src/lib/auth/__tests__/permissions.test.ts`

**Step 1: Add tests verifying old functions still work after refactor**

These tests already exist and should continue passing. No new RED tests needed — the existing tests for `canEdit`, `isAdmin`, etc. serve as regression. Skip to Task 4.

---

## Task 4: RED — Tests for consumer migration (page-level `can()` usage)

**Files:**
- Create: `frontend/src/lib/auth/__tests__/permission-matrix-integration.test.ts`

**Step 1: Write integration tests verifying real-world permission scenarios from page consumers**

```typescript
import { can, Resource, Action } from '../permissions'
import { UserRole } from '@/types/auth'

describe('permission matrix — page-level scenarios', () => {
  describe('announcements page (currently canEdit)', () => {
    it.each([
      [UserRole.SYSTEM_ADMIN, true],
      [UserRole.METHODIST, true],
      [UserRole.ACADEMIC_SECRETARY, true],
      [UserRole.TEACHER, true],
      [UserRole.STUDENT, false],
    ])('%s can create announcements: %s', (role, expected) => {
      // Announcements map to ASSIGNMENTS resource (they are a type of assignment/content)
      // Actually, announcements don't map neatly. They use the generic "canEdit" pattern.
      // For now, announcements follow: all EDIT_ROLES can create → map to assignments.create
      expect(can(role, Resource.ASSIGNMENTS, Action.CREATE)).toBe(expected)
    })
  })

  describe('documents page', () => {
    it('all roles except student can create documents', () => {
      expect(can(UserRole.SYSTEM_ADMIN, Resource.CURRICULUM, Action.CREATE)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.CURRICULUM, Action.CREATE)).toBe(true)
      expect(can(UserRole.STUDENT, Resource.CURRICULUM, Action.CREATE)).toBe(false)
    })
  })

  describe('users page — admin-only CRUD', () => {
    it('only admin can create/delete users', () => {
      expect(can(UserRole.SYSTEM_ADMIN, Resource.USERS, Action.CREATE)).toBe(true)
      expect(can(UserRole.SYSTEM_ADMIN, Resource.USERS, Action.DELETE)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.USERS, Action.CREATE)).toBe(false)
      expect(can(UserRole.TEACHER, Resource.USERS, Action.DELETE)).toBe(false)
    })

    it('non-admin roles can read users (limited)', () => {
      expect(can(UserRole.METHODIST, Resource.USERS, Action.READ)).toBe(true)
      expect(can(UserRole.TEACHER, Resource.USERS, Action.READ)).toBe(true)
    })
  })

  describe('schedule page', () => {
    it('secretary has full schedule control', () => {
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.SCHEDULE, Action.CREATE)).toBe(true)
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.SCHEDULE, Action.DELETE)).toBe(true)
    })

    it('teacher and student can only read schedule', () => {
      expect(can(UserRole.TEACHER, Resource.SCHEDULE, Action.READ)).toBe(true)
      expect(can(UserRole.TEACHER, Resource.SCHEDULE, Action.CREATE)).toBe(false)
      expect(can(UserRole.STUDENT, Resource.SCHEDULE, Action.READ)).toBe(true)
      expect(can(UserRole.STUDENT, Resource.SCHEDULE, Action.CREATE)).toBe(false)
    })
  })

  describe('reports page', () => {
    it('admin/methodist/secretary have full reports access', () => {
      expect(can(UserRole.SYSTEM_ADMIN, Resource.REPORTS, Action.CREATE)).toBe(true)
      expect(can(UserRole.METHODIST, Resource.REPORTS, Action.CREATE)).toBe(true)
      expect(can(UserRole.ACADEMIC_SECRETARY, Resource.REPORTS, Action.CREATE)).toBe(true)
    })

    it('teacher has limited (read only)', () => {
      expect(can(UserRole.TEACHER, Resource.REPORTS, Action.READ)).toBe(true)
      expect(can(UserRole.TEACHER, Resource.REPORTS, Action.CREATE)).toBe(false)
    })

    it('student denied', () => {
      expect(can(UserRole.STUDENT, Resource.REPORTS, Action.READ)).toBe(false)
    })
  })
})
```

**Step 2: Run to verify they fail**

Run: `cd frontend && npx jest src/lib/auth/__tests__/permission-matrix-integration.test.ts --no-coverage 2>&1 | tail -20`
Expected: Already PASS (since Task 2 implemented the matrix). These tests serve as documentation of the consumer migration contract.

**Step 3: Commit**

```bash
git add frontend/src/lib/auth/__tests__/permission-matrix-integration.test.ts
git commit -m "test(permissions): add integration tests for page-level permission scenarios

Verifies real-world permission patterns used by announcements, documents,
users, schedule, and reports pages against the resource-based matrix."
```

---

## Task 5: Migrate consumers — replace `canEdit(role)` with `can(role, resource, action)`

**Files to modify (6 pages):**
- `frontend/src/app/announcements/page.tsx` — line 40, 54
- `frontend/src/app/calendar/page.tsx` — line 20, 38
- `frontend/src/app/dashboard/page.tsx` — line 8, 55
- `frontend/src/app/documents/page.tsx` — line 30, 147
- `frontend/src/app/documents/templates/page.tsx` — line 18, 23
- `frontend/src/app/files/page.tsx` — line 34, 58
- `frontend/src/app/users/page.tsx` — line 57 (manual isAdmin check)

**Step 1: Update each page import + usage**

For each page, replace:
```typescript
import { canEdit } from '@/lib/auth/permissions'
// ...
const userCanEdit = canEdit(user?.role)
```

With the appropriate resource-specific check:
```typescript
import { can, Resource, Action } from '@/lib/auth/permissions'
// ...
const userCanEdit = can(user?.role, Resource.XXXX, Action.CREATE)
```

**Mapping:**
| Page | Resource | Primary action checked |
|------|----------|----------------------|
| `announcements/page.tsx` | `Resource.ASSIGNMENTS` | `Action.CREATE` |
| `calendar/page.tsx` | `Resource.SCHEDULE` | `Action.CREATE` |
| `dashboard/page.tsx` | `Resource.CURRICULUM` | `Action.CREATE` (generic "can user create content") |
| `documents/page.tsx` | `Resource.CURRICULUM` | `Action.CREATE` |
| `documents/templates/page.tsx` | `Resource.CURRICULUM` | `Action.CREATE` |
| `files/page.tsx` | `Resource.CURRICULUM` | `Action.CREATE` (file upload = document creation) |
| `users/page.tsx` | `Resource.USERS` | `Action.CREATE` / `Action.DELETE` |

**Special case — users/page.tsx:**
Replace line 57 `const isAdmin = currentUser?.role === UserRole.SYSTEM_ADMIN` with:
```typescript
import { can, Resource, Action } from '@/lib/auth/permissions'
// ...
const canManageUsers = can(currentUser?.role, Resource.USERS, Action.CREATE)
```
Then replace all `isAdmin` usages in that file with `canManageUsers`.

**Step 2: Run full test suite**

Run: `cd frontend && npx jest --no-coverage 2>&1 | tail -30`
Expected: ALL PASS (page-level tests mock the layout, permission functions are not directly tested there)

**Step 3: Commit**

```bash
git add frontend/src/app/announcements/page.tsx \
  frontend/src/app/calendar/page.tsx \
  frontend/src/app/dashboard/page.tsx \
  frontend/src/app/documents/page.tsx \
  frontend/src/app/documents/templates/page.tsx \
  frontend/src/app/files/page.tsx \
  frontend/src/app/users/page.tsx
git commit -m "feat(permissions): migrate 7 pages from canEdit() to can(role, resource, action)

Each page now uses resource-specific permission checks:
- announcements → assignments.create
- calendar → schedule.create
- dashboard → curriculum.create
- documents/templates → curriculum.create
- files → curriculum.create
- users → users.create/delete (replaces manual isAdmin check)"
```

---

## Task 6: Clean up — remove unused exports if safe

**Files:**
- Modify: `frontend/src/lib/auth/permissions.ts`

**Step 1: Check if `canEdit`, `canCreate`, `canDelete`, `isViewOnly` are still imported anywhere**

Run: `grep -r "canEdit\|canCreate\|canDelete\|isViewOnly\|VIEW_ONLY_ROLES" frontend/src/ --include="*.ts" --include="*.tsx" | grep -v __tests__ | grep -v permissions.ts`

If no production imports remain, mark old functions with `@deprecated` JSDoc (do NOT delete — components like DocumentList still receive `canEdit` as a prop name, and third-party code may depend on these).

**Step 2: Commit**

```bash
git add frontend/src/lib/auth/permissions.ts
git commit -m "refactor(permissions): deprecate legacy canEdit/isViewOnly functions

Old functions still exported for backward compat but marked @deprecated.
All page-level consumers now use can(role, resource, action)."
```

---

## Task 7: Version bump + release

**Step 1: Determine version**

This is a feature (new permission system) — bump minor: `0.103.0` → `0.104.0`.
Update all 8 version files per `memory/conventions.md`.

**Step 2: Commit + tag**

Standard release flow per project conventions.

---

## Summary

| Task | Type | What |
|------|------|------|
| 1 | RED | Tests for Resource/Action/AccessLevel enums + `can()` + `getAccessLevel()` |
| 2 | GREEN | Implement enums + permission matrix + `can()` + `getAccessLevel()` |
| 3 | — | Skipped (existing tests serve as regression) |
| 4 | TEST | Integration tests for page-level permission scenarios |
| 5 | FEAT | Migrate 7 pages from `canEdit()` to `can(role, resource, action)` |
| 6 | REFACTOR | Deprecate old functions |
| 7 | RELEASE | Version bump 0.104.0 |

Total: ~5 commits (2 TDD + 1 integration test + 1 migration + 1 deprecation) + release commit.
