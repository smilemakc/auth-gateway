import React from 'react';
import { useLanguage } from '../../services/i18n';
import type { CreateLDAPConfigRequest } from '@auth-gateway/client-sdk';

interface LDAPSearchFieldsProps {
  formData: CreateLDAPConfigRequest;
  onFormChange: (data: Partial<CreateLDAPConfigRequest>) => void;
}

const LDAPSearchFields: React.FC<LDAPSearchFieldsProps> = ({
  formData,
  onFormChange,
}) => {
  const { t } = useLanguage();

  return (
    <>
      <div className="border-b border-border pb-6">
        <h2 className="text-lg font-semibold text-foreground mb-4">{t('ldap_edit.user_search')}</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('ldap_edit.user_search_base')}</label>
            <input
              type="text"
              value={formData.user_search_base}
              onChange={(e) => onFormChange({ user_search_base: e.target.value })}
              className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="ou=users,dc=example,dc=com"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('ldap_edit.user_search_filter')}</label>
            <input
              type="text"
              value={formData.user_search_filter}
              onChange={(e) => onFormChange({ user_search_filter: e.target.value })}
              className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="(objectClass=person)"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('ldap_edit.user_id_attr')}</label>
            <input
              type="text"
              value={formData.user_id_attribute}
              onChange={(e) => onFormChange({ user_id_attribute: e.target.value })}
              className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="uid"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('ldap_edit.user_email_attr')}</label>
            <input
              type="text"
              value={formData.user_email_attribute}
              onChange={(e) => onFormChange({ user_email_attribute: e.target.value })}
              className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="mail"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('ldap_edit.user_name_attr')}</label>
            <input
              type="text"
              value={formData.user_name_attribute}
              onChange={(e) => onFormChange({ user_name_attribute: e.target.value })}
              className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="cn"
            />
          </div>
        </div>
      </div>

      <div className="border-b border-border pb-6">
        <h2 className="text-lg font-semibold text-foreground mb-4">{t('ldap_edit.group_search')}</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('ldap_edit.group_search_base')}</label>
            <input
              type="text"
              value={formData.group_search_base}
              onChange={(e) => onFormChange({ group_search_base: e.target.value })}
              className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="ou=groups,dc=example,dc=com"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('ldap_edit.group_search_filter')}</label>
            <input
              type="text"
              value={formData.group_search_filter}
              onChange={(e) => onFormChange({ group_search_filter: e.target.value })}
              className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="(objectClass=group)"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('ldap_edit.group_id_attr')}</label>
            <input
              type="text"
              value={formData.group_id_attribute}
              onChange={(e) => onFormChange({ group_id_attribute: e.target.value })}
              className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="cn"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('ldap_edit.group_name_attr')}</label>
            <input
              type="text"
              value={formData.group_name_attribute}
              onChange={(e) => onFormChange({ group_name_attribute: e.target.value })}
              className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="cn"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('ldap_edit.group_member_attr')}</label>
            <input
              type="text"
              value={formData.group_member_attribute}
              onChange={(e) => onFormChange({ group_member_attribute: e.target.value })}
              className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="member"
            />
          </div>
        </div>
      </div>
    </>
  );
};

export default LDAPSearchFields;
