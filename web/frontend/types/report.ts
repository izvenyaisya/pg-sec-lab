export interface InstanceInfo {
  version: string;
  settings: Record<string, string>;
}

export interface RoleInfo {
  name: string;
  login: boolean;
  superuser: boolean;
  bypassrls: boolean;
  grants: string[];
}

export interface TableInfo {
  schema: string;
  name: string;
  rls_enabled: boolean;
}

export type Severity = "info" | "warning" | "critical";

export interface Finding {
  severity: Severity;
  code: string;
  message: string;
}

export interface PolicyReport {
  instance: InstanceInfo;
  roles: RoleInfo[];
  tables: TableInfo[];
  findings: Finding[];
}
