

document.addEventListener('DOMContentLoaded', () => {
	let url = new URL(window.location);
	let sp = new URLSearchParams(url.search.slice(1));
	let state = sp.get('state');
	let code = sp.get('code');

	axios.post('//api.local.test/api/token-exchange', {
			state: state,
			code: code,
		}).then( (re) => {
			console.log(re.data);
			localStorage.setItem("token", re.data.token);
			localStorage.setItem("githubName", re.data.githubName);
			localStorage.setItem("expire", re.data.expire);
			location.href = "/";
		}).catch( (err) => {
			console.error(err);
			document.body.innerHTML = `
			<p>Error.Please retry.</p>
				<p><a href="/">back</a></p>
				`;
		});
}); 
