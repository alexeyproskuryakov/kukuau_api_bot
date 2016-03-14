Ext.define('Console.model.Profile', {
	extend: 'Ext.data.Model',
	fields: [
	'id', 
	'image_url', 
	'name', 
	'short_description', 
	'text_description', 
	'enable',
	'public'
	],
	associations: [{
		type: 'hasMany',
		model: 'Console.model.Contact',
		name: 'contacts'
	}, {
		type:'hasMany',
		model:'Console.model.Group',
		name:'groups'
	}],
	proxy: {
		type: 'ajax',
		api: {
			read: '/profile/read',
			create: '/profile/create',
			update: '/profile/update',
			destroy: '/profile/delete'
		}
	}
});
