
document.write("site.js loaded");

// vardump to element
function disp(element, obj) {
	if (typeof obj == "string") {
		element.textContent = obj;
	} else if(obj instanceof Error){
		element.textContent = obj.toString();
	} else {
		let str = JSON.stringify(obj);
		element.textContent = str;
	}
}

/**
 * Check login expiration
 */
let token = localStorage.getItem("token");
if(token){
	let expireStr = localStorage.getItem("expire");
	let expireDate = new Date(expireStr);
	let now = Date.now();
	let expire = expireDate.getTime();
	if(expire - now < 0){
		localStorage.removeItem("token");
	}
}


/**
 * call hello API GET, PUT and private method
 **/
document.addEventListener("DOMContentLoaded", () => {
	// for vardump
	let preGet = document.querySelector("pre.get");
	let prePut = document.querySelector("pre.put");

	// GET call.
	axios.get("//api.local.test/api/hello" 
	).then( re => {
		disp(preGet, re.data);
	}).catch(e => {
		disp(preGet, e);
	});

	// PUT call
	axios.put("//api.local.test/api/hello" 
	).then( re => {
		disp(prePut, re.data);
	}).catch(e => {
		disp(prePut, e);
	});

	// private
	let token = localStorage.getItem("token");
	if(token){
		let elmPrivate = document.querySelectorAll(".private");
		let elmPublic = document.querySelectorAll(".public");
		let prePrivate = document.querySelector(".private pre");

		for(let elm of elmPrivate){
			elm.style.visibility = "visible";
		}
		for(let elm of elmPublic){
			elm.style.visibility = "hidden";
		}
		axios.post("//api.local.test/api/private/hello",
			{token:token}
		).then( re => {
			disp(prePrivate, re.data);
		}).catch( e => {
			disp(prePrivate, e.response.data);
		});

		let name = localStorage.getItem("githubName");
		document.querySelector("#githubName").textContent = "hello " + name;
	}

	/**
	 * handle logout
	 */
	document.querySelector("#logout").onclick = function () {
		localStorage.removeItem("token");
		location.reload();
		return false;
	};
});

