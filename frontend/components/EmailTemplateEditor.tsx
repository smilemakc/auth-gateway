
import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Save, Eye, Code, RefreshCw, Check, Loader2 } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useEmailTemplateDetail, useUpdateEmailTemplate } from '../hooks/useEmailTemplates';

const EmailTemplateEditor: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();
  const [subject, setSubject] = useState('');
  const [bodyHtml, setBodyHtml] = useState('');
  const [saved, setSaved] = useState(false);
  const [activeTab, setActiveTab] = useState<'editor' | 'preview'>('editor');

  const { data: template, isLoading } = useEmailTemplateDetail(id || '');
  const updateMutation = useUpdateEmailTemplate();

  useEffect(() => {
    if (template) {
      setSubject(template.subject || '');
      setBodyHtml(template.html_body || '');
    }
  }, [template]);

  const handleSave = async () => {
    if (id) {
      try {
        await updateMutation.mutateAsync({
          id,
          data: { subject, html_body: bodyHtml }
        });
        setSaved(true);
        setTimeout(() => setSaved(false), 2000);
      } catch (err) {
        console.error('Failed to save template:', err);
      }
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  if (!template) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">Template not found</p>
      </div>
    );
  }

  // Simple interpolation for preview
  const getPreviewHtml = () => {
    let html = bodyHtml;
    // Mock variable replacement for preview
    const mockValues: Record<string, string> = {
      '{{name}}': 'John Doe',
      '{{email}}': 'john@example.com',
      '{{action_url}}': '#',
      '{{username}}': 'john_doe_99',
      '{{ip_address}}': '192.168.1.1',
      '{{os}}': 'Windows 11'
    };

    (template.variables || []).forEach(v => {
      const val = mockValues[v] || `[${v}]`;
      html = html.replace(new RegExp(v, 'g'), val);
    });

    return html;
  };

  return (
    <div className="h-[calc(100vh-8rem)] flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between mb-4 flex-shrink-0">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate('/settings/email-templates')}
            className="p-2 hover:bg-accent rounded-lg transition-colors text-muted-foreground"
          >
            <ArrowLeft size={24} />
          </button>
          <div>
            <h1 className="text-xl font-bold text-foreground">{template.name}</h1>
            <p className="text-xs text-muted-foreground">Edit template content</p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={() => setActiveTab(activeTab === 'editor' ? 'preview' : 'editor')}
            className="lg:hidden p-2 text-muted-foreground bg-card border border-border rounded-lg"
          >
             {activeTab === 'editor' ? <Eye size={20} /> : <Code size={20} />}
          </button>
          <button
            onClick={handleSave}
            disabled={updateMutation.isPending}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg font-medium text-sm transition-colors
              ${saved
                ? 'bg-success text-primary-foreground'
                : 'bg-primary text-primary-foreground hover:bg-primary-600'}`}
          >
            {updateMutation.isPending ? <RefreshCw size={18} className="animate-spin" /> :
             saved ? <Check size={18} /> : <Save size={18} />}
            {saved ? t('common.saved') : t('common.save')}
          </button>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex gap-6 min-h-0">

        {/* Editor Pane */}
        <div className={`flex-1 flex flex-col gap-4 ${activeTab === 'preview' ? 'hidden lg:flex' : 'flex'}`}>
          <div className="bg-card p-4 rounded-xl shadow-sm border border-border flex-shrink-0">
            <label className="block text-sm font-medium text-foreground mb-1">{t('email.subject')}</label>
            <input
              type="text"
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              className="w-full px-3 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent outline-none"
            />
            <div className="mt-3 flex flex-wrap gap-2">
              <span className="text-xs text-muted-foreground py-1">{t('email.vars')}:</span>
              {(template.variables || []).map(v => (
                <button
                  key={v}
                  onClick={() => setBodyHtml(prev => prev + v)}
                  className="text-xs bg-muted hover:bg-accent text-foreground px-2 py-1 rounded font-mono border border-border transition-colors"
                  title="Click to insert"
                >
                  {v}
                </button>
              ))}
            </div>
          </div>

          <div className="flex-1 bg-card rounded-xl shadow-sm border border-border flex flex-col overflow-hidden">
            <div className="bg-muted px-4 py-2 border-b border-border flex items-center justify-between">
              <span className="text-xs font-semibold text-muted-foreground uppercase">{t('email.body')}</span>
            </div>
            <textarea
              value={bodyHtml}
              onChange={(e) => setBodyHtml(e.target.value)}
              className="flex-1 w-full p-4 font-mono text-sm outline-none resize-none"
              spellCheck="false"
            />
          </div>
        </div>

        {/* Preview Pane */}
        <div className={`flex-1 flex flex-col ${activeTab === 'editor' ? 'hidden lg:flex' : 'flex'}`}>
          <div className="flex-1 bg-card rounded-xl shadow-sm border border-border overflow-hidden flex flex-col">
            <div className="bg-muted px-4 py-2 border-b border-border flex items-center justify-between">
              <span className="text-xs font-semibold text-muted-foreground uppercase">{t('email.preview')}</span>
              <span className="text-xs text-muted-foreground">Values are mocked</span>
            </div>
            <div className="bg-muted p-4 border-b border-border">
               <div className="text-sm font-medium text-muted-foreground mb-1">{t('email.subject')}:</div>
               <div className="text-foreground font-medium">{subject}</div>
            </div>
            <div className="flex-1 bg-card relative">
              <iframe
                title="preview"
                srcDoc={getPreviewHtml()}
                className="absolute inset-0 w-full h-full border-0"
                sandbox="allow-same-origin"
              />
            </div>
          </div>
        </div>

      </div>
    </div>
  );
};

export default EmailTemplateEditor;
