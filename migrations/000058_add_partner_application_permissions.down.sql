DELETE FROM permissions WHERE name IN
    ('partner_application.apply', 'partner_application.review', 'partner_application.manage_all');
