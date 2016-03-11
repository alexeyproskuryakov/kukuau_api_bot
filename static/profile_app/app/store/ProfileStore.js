Ext.define('Console.store.ProfileStore', {
    extend: 'Ext.data.Store',
    model: 'Console.model.Profile',
    storeId: 'ProfileStore',
    proxy: {
        type: 'ajax',
        url: '/profile/data',
        reader: {
            type: 'json',
            root: 'profiles',
            successProperty: 'success'
        }
    }
});
